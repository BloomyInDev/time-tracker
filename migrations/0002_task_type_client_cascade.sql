-- 0002_task_type_client_cascade: foreign keys are enforced from this release
-- on, and the join table between clients and task types is the one place where
-- a delete should take its rows with it rather than be refused. Deleting a
-- client or a task type must not be blocked by a checkbox someone ticked on an
-- edit page; the rows carry no information of their own.
--
-- Tasks are deliberately left out of this: their foreign keys stay RESTRICT so
-- deleting a client can't silently destroy the time logged against it. That
-- refusal surfaces as db.ErrInUse and, in the UI, as a 409.
--
-- SQLite can't ALTER a constraint, so the table is rebuilt. Nothing references
-- it, which is what makes the drop-and-rename safe to do in place.

CREATE TABLE task_types_for_client_new (
    client_id    INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    task_type_id INTEGER NOT NULL REFERENCES task_types(id) ON DELETE CASCADE,
    PRIMARY KEY (client_id, task_type_id)
);

INSERT INTO task_types_for_client_new (client_id, task_type_id)
SELECT client_id, task_type_id FROM task_types_for_client;

DROP TABLE task_types_for_client;

ALTER TABLE task_types_for_client_new RENAME TO task_types_for_client;
