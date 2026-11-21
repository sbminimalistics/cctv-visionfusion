create table
    captured_events (
        id uuid primary key,
        src text,
        src_type text,
        type text,
        created_at timestamptz NOT NULL DEFAULT now (),
        trace_id text
    );

create table
    detections (
        id bigserial primary key,
        event_id uuid references captured_events (id) on delete cascade,
        model_name text NOT NULL,
        seaweed_object_key text NOT NULL,
        image_width smallint NOT NULL,
        image_height smallint NOT NULL,
        trace_id varchar(32),
        detected_at timestamptz NOT NULL,
        created_at timestamptz NOT NULL DEFAULT now ()
    );

create table
    detection_objects (
        id bigserial PRIMARY KEY,
        detection_id integer NOT NULL REFERENCES detections (id) ON DELETE CASCADE,
        class_id smallint NOT NULL,
        class_name text NOT NULL,
        x smallint NOT NULL,
        y smallint NOT NULL,
        width smallint NOT NULL,
        height smallint NOT NULL,
        confidence real NOT NULL
    );

CREATE UNIQUE INDEX detections_idx ON detections (id);

CREATE INDEX idx_events_created_at ON captured_events (created_at DESC);

CREATE INDEX idx_events_trace_id ON captured_events (trace_id)
WHERE
    trace_id IS NOT NULL;

CREATE INDEX idx_detections_event_id ON detections (event_id);

CREATE INDEX idx_detection_objects_detection_id ON detection_objects (id);