-- Write your migrate up statements here

-- TODO: Add validation
-- TODO: represent graph

CREATE TABLE type_references
(
    id uuid NOT NULL PRIMARY KEY,
    path     varchar(256),
    revision varchar(256)
);

CREATE TABLE typeinstances
(
    id uuid NOT NULL PRIMARY KEY,
    type_ref uuid NOT NULL,
    FOREIGN KEY (type_ref) REFERENCES type_references(id)
);

CREATE TABLE typeinstance_resource_versions
(
    id uuid NOT NULL PRIMARY KEY,
    typeinstance_id uuid NOT NULL,
    resource_version int NOT NULL,
    FOREIGN KEY (typeinstance_id) REFERENCES typeinstances(id),
    value            jsonb,
    UNIQUE (typeinstance_id, resource_version)
);

---- create above / drop below ----

-- Write your down migrate statements here. If this migration is irreversible
-- then delete the separator line above.


