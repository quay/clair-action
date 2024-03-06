package datastore

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	sqldriver "database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	version "github.com/hashicorp/go-version"
	"github.com/quay/clair-action/migrations"
	"github.com/quay/claircore"
	"github.com/quay/claircore/datastore"
	"github.com/quay/claircore/libvuln/driver"
	"github.com/quay/zlog"
	"github.com/remind101/migrate"
	"modernc.org/sqlite"
)

type intVersion [10]int32

func (v *intVersion) String() string {
	var strs []string
	for _, p := range v {
		strs = append(strs, fmt.Sprint(p))
	}
	return strings.Join(strs, ".")
}

func (v *intVersion) FromString(str string) error {
	str = strings.Trim(str, "{}")
	sl := strings.Split(str, ",")
	for i := 0; i < len(sl); i++ {
		p, err := strconv.ParseInt(sl[i], 10, 32)
		if err != nil {
			return err
		}
		v[i] = int32(p)
	}
	return nil
}

func NewSQLiteMatcherStore(DSN string, doMigration bool) (*sqliteMatcherStore, error) {
	sqlite.MustRegisterDeterministicScalarFunction("version_in", 2, _sqliteVersionIn)
	db, err := sql.Open("sqlite", DSN)
	if err != nil {
		return nil, err
	}

	if doMigration {
		migrator := migrate.NewMigrator(db)
		migrator.Table = migrations.MigrationTable
		if err := migrator.Exec(migrate.Up, migrations.MatcherMigrations...); err != nil {
			return nil, err
		}
	}
	return &sqliteMatcherStore{conn: db}, nil
}

// _sqliteVersionIn is registered and used to determine if a package version falls within a version
// range, the lower bound is considered inclusive and the upper is considered exclusive.

// vulnerable range is expected as a pair of 10 part, comma seperated version representations
// separated by `__` e.g. "{0,0,0,0,0,0,0,0,0,0}__{3,6,2147483647,0,0,0,0,0,0,0}"
func _sqliteVersionIn(ctx *sqlite.FunctionContext, args []sqldriver.Value) (sqldriver.Value, error) {
	if len(args) != 2 {
		return nil, errors.New("version_in must be passed 2 args")
	}
	pkgVer, ok := args[0].(string)
	if !ok {
		return nil, errors.New("could not convert package version arg to string")
	}
	vulnRange, ok := args[1].(string)
	if !ok {
		return nil, errors.New("could not convert vulnerable range arg to string")
	}

	var lower, upper intVersion
	ver, err := version.NewVersion(pkgVer)
	if err != nil {
		return false, fmt.Errorf("could not create version: %v", err)
	}

	vers := strings.Split(vulnRange, "__")
	if len(vers) != 2 {
		return false, fmt.Errorf("invalid version range %s", vulnRange)
	}
	err = lower.FromString(vers[0])
	if err != nil {
		return false, fmt.Errorf("could not create lower version: %v", err)
	}
	err = upper.FromString(vers[1])
	if err != nil {
		return false, fmt.Errorf("could not create upper version: %v", err)
	}
	constraints, err := version.NewConstraint(">=" + lower.String() + ", < " + upper.String())
	if err != nil {
		return false, fmt.Errorf("could not compare versions: %v", err)
	}
	if constraints.Check(ver) {
		return true, nil
	}
	return false, nil
}

type sqliteMatcherStore struct {
	conn *sql.DB
}

// UpdateEnrichments creates a new EnrichmentUpdateOperation, inserts the provided
// EnrichmentRecord(s), and ensures enrichments from previous updates are not
// queries by clients.
func (ms *sqliteMatcherStore) UpdateEnrichments(ctx context.Context, updaterName string, fp driver.Fingerprint, es []driver.EnrichmentRecord) (uuid.UUID, error) {
	const (
		insert = `
	INSERT
	INTO
		enrichment (hash_kind, hash, updater, tags, data)
	VALUES
		($1, $2, $3, json_array($4), $5)
	ON CONFLICT
		(hash_kind, hash)
	DO
		NOTHING;`
	)

	var ref = uuid.New()
	tx, err := ms.conn.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()

	for i := range es {
		hashKind, hash := hashEnrichment(&es[i])
		_, err := tx.ExecContext(ctx, insert,
			hashKind, hash, updaterName, strings.Join(es[i].Tags, ","), es[i].Enrichment,
		)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to insert enrichment: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ref, nil
}

func hashEnrichment(r *driver.EnrichmentRecord) (k string, d []byte) {
	h := md5.New()
	sort.Strings(r.Tags)
	for _, t := range r.Tags {
		io.WriteString(h, t)
		h.Write([]byte("\x00"))
	}
	h.Write(r.Enrichment)
	return "md5", h.Sum(nil)
}

// UpdateVulnerabilities creates a new UpdateOperation, inserts the provided
// vulnerabilities, and ensures vulnerabilities from previous updates are
// not queried by clients.
func (ms *sqliteMatcherStore) UpdateVulnerabilities(ctx context.Context, updaterName string, fp driver.Fingerprint, vs []*claircore.Vulnerability) (uuid.UUID, error) {
	const (
		// Insert attempts to create a new vulnerability. It fails silently.
		insert = `
		INSERT INTO vuln (
			hash_kind, hash,
			name, updater, description, issued, links, severity, normalized_severity,
			package_name, package_version, package_module, package_arch, package_kind,
			dist_id, dist_name, dist_version, dist_version_code_name, dist_version_id, dist_arch, dist_cpe, dist_pretty_name,
			repo_name, repo_key, repo_uri,
			fixed_in_version, arch_operation, version_kind, vulnerable_range
		) VALUES (
		  $1, $2,
		  $3, $4, $5, $6, $7, $8, $9,
		  $10, $11, $12, $13, $14,
		  $15, $16, $17, $18, $19, $20, $21, $22,
		  $23, $24, $25,
		  $26, $27, $28, $29
		)
		ON CONFLICT (hash_kind, hash) DO NOTHING;`
	)

	var ref = uuid.New()
	tx, err := ms.conn.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()

	for _, vuln := range vs {
		if vuln.Package == nil || vuln.Package.Name == "" {
			continue
		}

		pkg := vuln.Package
		dist := vuln.Dist
		repo := vuln.Repo
		if dist == nil {
			dist = &claircore.Distribution{}
		}
		if repo == nil {
			repo = &claircore.Repository{}
		}
		hashKind, hash := md5Vuln(vuln)
		vKind, vrLower, vrUpper := rangefmt(vuln.Range)

		if _, err := tx.ExecContext(ctx, insert,
			hashKind, hash,
			vuln.Name, vuln.Updater, vuln.Description, vuln.Issued.Format(time.RFC3339), vuln.Links, vuln.Severity, vuln.NormalizedSeverity,
			pkg.Name, pkg.Version, pkg.Module, pkg.Arch, pkg.Kind,
			dist.DID, dist.Name, dist.Version, dist.VersionCodeName, dist.VersionID, dist.Arch, &dist.CPE, dist.PrettyName,
			repo.Name, repo.Key, repo.URI,
			vuln.FixedInVersion, vuln.ArchOperation, vKind, strings.Join([]string{vrLower, vrUpper}, "__"),
		); err != nil {
			return uuid.Nil, fmt.Errorf("failed to insert vulnerability: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ref, nil
}

// Md5Vuln creates an md5 hash from the members of the passed-in Vulnerability,
// giving us a stable, context-free identifier for this revision of the
// Vulnerability.
func md5Vuln(v *claircore.Vulnerability) (string, []byte) {
	var b bytes.Buffer
	b.WriteString(v.Name)
	b.WriteString(v.Description)
	b.WriteString(v.Issued.String())
	b.WriteString(v.Links)
	b.WriteString(v.Severity)
	if v.Package != nil {
		b.WriteString(v.Package.Name)
		b.WriteString(v.Package.Version)
		b.WriteString(v.Package.Module)
		b.WriteString(v.Package.Arch)
		b.WriteString(v.Package.Kind)
	}
	if v.Dist != nil {
		b.WriteString(v.Dist.DID)
		b.WriteString(v.Dist.Name)
		b.WriteString(v.Dist.Version)
		b.WriteString(v.Dist.VersionCodeName)
		b.WriteString(v.Dist.VersionID)
		b.WriteString(v.Dist.Arch)
		b.WriteString(v.Dist.CPE.BindFS())
		b.WriteString(v.Dist.PrettyName)
	}
	if v.Repo != nil {
		b.WriteString(v.Repo.Name)
		b.WriteString(v.Repo.Key)
		b.WriteString(v.Repo.URI)
	}
	b.WriteString(v.ArchOperation.String())
	b.WriteString(v.FixedInVersion)
	if k, l, u := rangefmt(v.Range); k != nil {
		b.WriteString(*k)
		b.WriteString(l)
		b.WriteString(u)
	}
	s := md5.Sum(b.Bytes())
	return "md5", s[:]
}

func rangefmt(r *claircore.Range) (kind *string, lower, upper string) {
	lower, upper = "{}", "{}"
	if r == nil || r.Lower.Kind != r.Upper.Kind {
		return kind, lower, upper
	}

	kind = &r.Lower.Kind // Just tested the both kinds are the same.
	v := &r.Lower
	var buf strings.Builder
	b := make([]byte, 0, 16) // 16 byte wide scratch buffer

	buf.WriteByte('{')
	for i := 0; i < 10; i++ {
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.Write(strconv.AppendInt(b, int64(v.V[i]), 10))
	}
	buf.WriteByte('}')
	lower = buf.String()
	buf.Reset()
	v = &r.Upper
	buf.WriteByte('{')
	for i := 0; i < 10; i++ {
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.Write(strconv.AppendInt(b, int64(v.V[i]), 10))
	}
	buf.WriteByte('}')
	upper = buf.String()

	return kind, lower, upper
}

// Initialized reports whether the vulnstore contains vulnerabilities.
func (ms *sqliteMatcherStore) Initialized(_ context.Context) (bool, error) {
	return true, nil
}

// get finds the vulnerabilities which match each package provided in the packages array
// this maybe a one to many relationship. each package is assumed to have an ID.
// a map of Package.ID => Vulnerabilities is returned.
func (ms *sqliteMatcherStore) Get(ctx context.Context, records []*claircore.IndexRecord, opts datastore.GetOpts) (map[string][]*claircore.Vulnerability, error) {
	tx, err := ms.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	results := make(map[string][]*claircore.Vulnerability)
	vulnSet := make(map[string]map[string]struct{})

	for _, record := range records {
		query, err := buildGetQuery(record, &opts)
		if err != nil {
			// if we cannot build a query for an individual record continue to the next
			zlog.Debug(ctx).
				Err(err).
				Str("record", fmt.Sprintf("%+v", record)).
				Msg("could not build query for record")
			continue
		}
		// queue the select query
		rows, err := tx.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// unpack all returned rows into claircore.Vulnerability structs
		for rows.Next() {
			// fully allocate vuln struct
			v := &claircore.Vulnerability{
				Package: &claircore.Package{},
				Dist:    &claircore.Distribution{},
				Repo:    &claircore.Repository{},
			}

			var issued string
			var hashBin string
			err := rows.Scan(
				&hashBin,
				&v.Name,
				&v.Description,
				&issued,
				&v.Links,
				&v.Severity,
				&v.NormalizedSeverity,
				&v.Package.Name,
				&v.Package.Version,
				&v.Package.Module,
				&v.Package.Arch,
				&v.Package.Kind,
				&v.Dist.DID,
				&v.Dist.Name,
				&v.Dist.Version,
				&v.Dist.VersionCodeName,
				&v.Dist.VersionID,
				&v.Dist.Arch,
				&v.Dist.CPE,
				&v.Dist.PrettyName,
				&v.ArchOperation,
				&v.Repo.Name,
				&v.Repo.Key,
				&v.Repo.URI,
				&v.FixedInVersion,
				&v.Updater,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan vulnerability: %v", err)
			}

			v.ID = base64.StdEncoding.EncodeToString([]byte(hashBin))

			v.Issued, err = time.Parse(time.RFC3339, issued)
			if err != nil {
				return nil, fmt.Errorf("failed parse issued date: %v", err)
			}

			rid := record.Package.ID
			if _, ok := vulnSet[rid]; !ok {
				vulnSet[rid] = make(map[string]struct{})
			}
			if _, ok := vulnSet[rid][v.ID]; !ok {
				vulnSet[rid][v.ID] = struct{}{}
				results[rid] = append(results[rid], v)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %v", err)
	}
	return results, nil
}

func makePlaceholders(startIndex, length int) string {
	str := ""
	for i := startIndex; i < length+startIndex; i++ {
		str = str + fmt.Sprintf("$%d,", i)
	}
	return "(" + strings.TrimRight(str, ",") + ")"
}

func (ms *sqliteMatcherStore) GetEnrichment(ctx context.Context, kind string, tags []string) ([]driver.EnrichmentRecord, error) {
	var query = `
	SELECT
		e.tags, e.data
	FROM
		enrichment AS e,
		json_each(e.tags)
	WHERE
		json_each.value IN ` + makePlaceholders(2, len(tags)) + ";"

	tx, err := ms.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	results := make([]driver.EnrichmentRecord, 0, 8) // Guess at capacity.
	args := []interface{}{kind}
	for _, v := range tags {
		args = append(args, v)
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		results = append(results, driver.EnrichmentRecord{})
		r := &results[i]
		var tags = []byte{}
		if err := rows.Scan(&tags, &r.Enrichment); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(tags, &r.Tags); err != nil {
			return nil, err
		}
		i++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// GetUpdateOperations returns a list of UpdateOperations in date descending
// order for the given updaters.
//
// The returned map is keyed by Updater implementation's unique names.
//
// If no updaters are specified, all UpdateOperations are returned.
func (ms *sqliteMatcherStore) GetUpdateOperations(ctx context.Context, kind driver.UpdateKind, updaters ...string) (map[string][]driver.UpdateOperation, error) {
	return nil, nil
}

// GetLatestUpdateRefs reports the latest update reference for every known
// updater.
func (ms *sqliteMatcherStore) GetLatestUpdateRefs(context.Context, driver.UpdateKind) (map[string][]driver.UpdateOperation, error) {
	panic("not implemented") // TODO: Implement
}

// GetLatestUpdateRef reports the latest update reference of any known
// updater.
func (ms *sqliteMatcherStore) GetLatestUpdateRef(context.Context, driver.UpdateKind) (uuid.UUID, error) {
	panic("not implemented") // TODO: Implement
}

// GetUpdateOperationDiff reports the UpdateDiff of the two referenced
// Operations.
//
// In diff(1) terms, this is like
//
//	diff prev cur
func (ms *sqliteMatcherStore) GetUpdateDiff(context.Context, uuid.UUID, uuid.UUID) (*driver.UpdateDiff, error) {
	return nil, nil
}

// GC stuff
// DeleteUpdateOperations removes an UpdateOperation.
// A call to GC must be run after this to garbage collect vulnerabilities associated
// with the UpdateOperation.
//
// The number of UpdateOperations deleted is returned.
func (ms *sqliteMatcherStore) DeleteUpdateOperations(context.Context, ...uuid.UUID) (int64, error) {
	return 0, nil
}

// GC will delete any update operations for an updater which exceeds the provided keep
// value.
//
// Implementations may throttle the GC process for datastore efficiency reasons.
//
// The returned int64 value indicates the remaining number of update operations needing GC.
// Running this method till the returned value is 0 accomplishes a full GC of the vulnstore.
func (ms *sqliteMatcherStore) GC(context.Context, int) (int64, error) {
	return 0, nil
}

// RecordUpdaterStatus records that an updater is up to date with vulnerabilities at this time
func (ms *sqliteMatcherStore) RecordUpdaterStatus(context.Context, string, time.Time, driver.Fingerprint, error) error {
	return nil
}

// RecordUpdaterSetStatus records that all updaters from an updater set are up to date with vulnerabilities at this time
func (ms *sqliteMatcherStore) RecordUpdaterSetStatus(context.Context, string, time.Time) error {
	return nil
}

// RecordUpdaterSetStatus records that all updaters from an updater set are up to date with vulnerabilities at this time
func (ms *sqliteMatcherStore) DeltaUpdateVulnerabilities(context.Context, string, driver.Fingerprint, []*claircore.Vulnerability, []string) (uuid.UUID, error) {
	panic("not implemented") // TODO: Implement when VEX updater is merged
}
