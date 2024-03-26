CREATE TABLE IF NOT EXISTS metadata (
	id          INTEGER PRIMARY KEY,
	kind        TEXT NOT NULL,
	hash_kind   TEXT NOT NULL,
	hash        TEXT NOT NULL,
	value       TEXT NOT NULL,
	UNIQUE (kind, hash_kind, hash)
);

DROP TABLE vuln;

CREATE TABLE vuln (
	id                     INTEGER PRIMARY KEY,
	hash_kind              TEXT NOT NULL,
	hash                   TEXT NOT NULL,
	updater                TEXT,
	name_id                INTEGER,
	description_id         INTEGER,
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
	UNIQUE (hash_kind, hash),
	FOREIGN KEY (description_id) REFERENCES metadata(id),
	FOREIGN KEY (name_id) REFERENCES metadata(id)
);

CREATE INDEX vuln_lookup_idx on vuln (package_name, dist_id,
                                         dist_name, dist_pretty_name,
                                         dist_version, dist_version_id,
                                         package_module, dist_version_code_name,
                                         repo_name, dist_arch,
                                         dist_cpe, repo_key,
                                         repo_uri);
CREATE INDEX vuln_lookup_updater ON vuln (updater);

