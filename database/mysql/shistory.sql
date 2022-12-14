CREATE TABLE `%s` (
    `rank`              bigint UNSIGNED     NOT NULL    AUTO_INCREMENT,
    `mode`              char(3)             NOT NULL,
    `version`           varchar(255)        NOT NULL,
    `script_name`       varchar(255)        NOT NULL,
    `description`       varchar(255)        NOT NULL    DEFAULT '',
    `checksum`          char(32)            NOT NULL    DEFAULT '',
    `applied_by`        varchar(255)        NOT NULL,
    `applied_at`        bigint UNSIGNED     NOT NULL,
    `execution_time`    int                 NOT NULL    DEFAULT 0,
    `success`           tinyint(1)          NOT NULL    DEFAULT 0,
    PRIMARY KEY(`rank`),
    INDEX (`mode`, `version`)
)
