-- update_operation
CREATE TABLE IF NOT EXISTS update_operation (
	id			INTEGER PRIMARY KEY,
	ref         TEXT UNIQUE,
	updater		TEXT NOT NULL,
	fingerprint TEXT,
	date		DATETIME DEFAULT CURRENT_TIMESTAMP,
	kind 		TEXT DEFAULT "vulnerability"
);

CREATE INDEX IF NOT EXISTS uo_lookup_updater ON update_operation (updater);
CREATE INDEX IF NOT EXISTS uo_lookup_kind ON update_operation (kind);

-- vuln
CREATE TABLE IF NOT EXISTS vuln (
	id                     INTEGER PRIMARY KEY,
	hash_kind              TEXT NOT NULL,
	hash                   TEXT NOT NULL,
	updater                TEXT,
	name                   TEXT,
	description            TEXT,
	issued                 TEXT,
	links                  TEXT,
	severity               TEXT,
	normalized_severity    TEXT,
	package_name           TEXT,
	package_version        TEXT,
	package_module         TEXT,
	package_arch           TEXT,
	package_kind           TEXT,
	dist_id                TEXT,
	dist_name              TEXT,
	dist_version           TEXT,
	dist_version_code_name TEXT,
	dist_version_id        TEXT,
	dist_arch              TEXT,
	dist_cpe               TEXT,
	dist_pretty_name       TEXT,
	repo_name              TEXT,
	repo_key               TEXT,
	repo_uri               TEXT,
	fixed_in_version       TEXT,
	arch_operation         TEXT,
	vulnerable_range       TEXT,
	version_kind           TEXT,
	UNIQUE (hash_kind, hash)
);

CREATE INDEX vuln_lookup_idx on vuln (package_name, dist_id,
                                         dist_name, dist_pretty_name,
                                         dist_version, dist_version_id,
                                         package_module, dist_version_code_name,
                                         repo_name, dist_arch,
                                         dist_cpe, repo_key,
                                         repo_uri);
CREATE INDEX vuln_lookup_updater ON vuln (updater);

-- uo_vuln                            
CREATE TABLE IF NOT EXISTS uo_vuln (
	uo   INTEGER REFERENCES update_operation (id) ON DELETE CASCADE,
	vuln INTEGER REFERENCES vuln             (id) ON DELETE CASCADE,
	PRIMARY KEY (uo, vuln)
);

CREATE INDEX uovuln_lookup_vuln ON uo_vuln (vuln);
CREATE INDEX uovuln_lookup_uo ON uo_vuln (uo);

-- enrichment
CREATE TABLE enrichment (
    id        INTEGER PRIMARY KEY,
    hash_kind TEXT,
    hash      TEXT,
    updater   TEXT,
    tags      TEXT,
    data      BLOB
);
CREATE UNIQUE INDEX enrichment_lookup_hash_kind_hash ON enrichment (hash_kind, hash);

-- uo_enrich
CREATE TABLE IF NOT EXISTS uo_enrich (
    uo          INTEGER REFERENCES update_operation (id) ON DELETE CASCADE,
    enrich      INTEGER REFERENCES enrichment (id) ON DELETE CASCADE,
    updater     TEXT,
    fingerprint TEXT,
    date        DATETIME DEFAULT CURRENT_TIMESTAMP
);
