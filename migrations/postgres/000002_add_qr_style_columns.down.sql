ALTER TABLE qr_codes
    DROP COLUMN IF EXISTS dot_style,
    DROP COLUMN IF EXISTS corner_style,
    DROP COLUMN IF EXISTS size,
    DROP COLUMN IF EXISTS margin;
