
-- SNIPPETS TABLE: stores snippet metadata
CREATE TABLE snippets (
    -- Internal ID (primary key)
    id SERIAL PRIMARY KEY,
    
    -- Public facing ID with Google Meet-style format (abc-defg-hij)
    public_id VARCHAR(12) NOT NULL UNIQUE,
    
    -- Snippet title (optional)
    title VARCHAR(255),
    
    -- Creation timestamp
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Expiration date (NULL means never expires)
    expires_at TIMESTAMP WITH TIME ZONE,
    
   
    -- Password hash stored using bcrypt/argon2 
    password_hash VARCHAR(255),

    edit_token VARCHAR(64) NOT NULL,

    view_count INTEGER NOT NULL DEFAULT 0,

    last_edited_at TIMESTAMP WITH TIME ZONE
);

-- SNIPPET_CONTENTS TABLE: stores the encrypted content and related encryption data
CREATE TABLE snippet_contents (
    -- Reference to the snippet this content belongs to
    snippet_id INTEGER NOT NULL REFERENCES snippets(id) ON DELETE CASCADE,
    
    -- The content type/language 
    content_type VARCHAR(100) NOT NULL DEFAULT 'text/plain',
    
    -- The encrypted content of the snippet 
    encrypted_content BYTEA NOT NULL,
    
    PRIMARY KEY (snippet_id)
);

-- Index to quickly find snippets by their public ID
CREATE INDEX idx_snippets_public_id ON snippets(public_id);

-- Index to efficiently query for expired snippets
CREATE INDEX idx_snippets_expires_at ON snippets(expires_at) 
WHERE expires_at IS NOT NULL;

-- Function to generate a Google Meet-style ID (abc-defg-hij)
CREATE OR REPLACE FUNCTION generate_snippet_id() 
RETURNS VARCHAR AS $$
DECLARE
    chars TEXT := 'abcdefghijkmnopqrstuvwxyz23456789';
    result VARCHAR := '';
    i INTEGER := 0;
BEGIN
    -- First segment (3 chars)
    FOR i IN 1..3 LOOP
        result := result || substr(chars, floor(random() * length(chars))::integer + 1, 1);
    END LOOP;
    
    result := result || '-';
    
    -- Second segment (4 chars)
    FOR i IN 1..4 LOOP
        result := result || substr(chars, floor(random() * length(chars))::integer + 1, 1);
    END LOOP;
    
    result := result || '-';
    
    -- Third segment (3 chars)
    FOR i IN 1..3 LOOP
        result := result || substr(chars, floor(random() * length(chars))::integer + 1, 1);
    END LOOP;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;


-- Trigger to generate a unique public_id for new snippets
CREATE OR REPLACE FUNCTION set_snippet_public_id()
RETURNS TRIGGER AS $$
DECLARE
    new_id VARCHAR;
    id_exists BOOLEAN;
BEGIN
    -- Generate IDs until we find one that doesn't exist
    LOOP
        new_id := generate_snippet_id();
        
        -- Check if this ID already exists
        SELECT EXISTS(SELECT 1 FROM snippets WHERE public_id = new_id) INTO id_exists;
        
        -- If the ID doesn't exist, use it
        IF NOT id_exists THEN
            NEW.public_id := new_id;
            EXIT;
        END IF;
    END LOOP;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_snippet_public_id
BEFORE INSERT ON snippets
FOR EACH ROW
WHEN (NEW.public_id IS NULL)
EXECUTE FUNCTION set_snippet_public_id();
