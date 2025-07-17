-- понять значения
CREATE TABLE IF NOT EXISTS todo_items (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    done BOOLEAN DEFAULT FALSE
);
