ALTER TABLE
    instance
SET
    (
        autovacuum_analyze_scale_factor = 0,
        autovacuum_analyze_threshold = 3666,
        autovacuum_vacuum_scale_factor = 0,
        autovacuum_vacuum_threshold = 7333
    );


ALTER TABLE
    instance_application
SET
    (
        autovacuum_analyze_scale_factor = 0,
        autovacuum_analyze_threshold = 3666,
        autovacuum_vacuum_scale_factor = 0,
        autovacuum_vacuum_threshold = 7333
    );

ALTER TABLE
    activity
SET
    (
        autovacuum_analyze_scale_factor = 0,
        autovacuum_analyze_threshold = 10000,
        autovacuum_vacuum_scale_factor = 0,
        autovacuum_vacuum_threshold = 20000
    );

ALTER TABLE
    event
SET
    (
        autovacuum_analyze_scale_factor = 0,
        autovacuum_analyze_threshold = 950,
        autovacuum_vacuum_scale_factor = 0,
        autovacuum_vacuum_threshold = 1900
    );

ALTER TABLE
    instance_status_history
SET
    (
        autovacuum_analyze_scale_factor = 0,
        autovacuum_analyze_threshold = 12500,
        autovacuum_vacuum_scale_factor = 0,
        autovacuum_vacuum_threshold = 25000
    );
