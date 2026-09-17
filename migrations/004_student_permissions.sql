-- 004_student_permissions.sql

-- 1. Tambahkan permission baru terkait student
INSERT INTO permissions (name, description) VALUES
    ('student:list',       'Melihat daftar seluruh mahasiswa'),
    ('student:read:any',   'Melihat data mahasiswa mana pun'),
    ('student:create',     'Menambahkan data mahasiswa'),
    ('student:update:any', 'Mengubah data mahasiswa mana pun'),
    ('student:delete',     'Menghapus data mahasiswa')
ON CONFLICT (name) DO NOTHING;

-- 2. Petakan permission ke role (admin & staff)
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT DO NOTHING;

-- 3. Tambahkan kolom owner_id pada tabel students
ALTER TABLE students 
    ADD COLUMN IF NOT EXISTS owner_id INT;

-- Amankan baris data lama (jika ada) agar mengarah ke user ID 1 (misal: admin)
UPDATE students SET owner_id = 1 WHERE owner_id IS NULL;

-- Ubah kolom menjadi NOT NULL dan beri Foreign Key ke tabel users
ALTER TABLE students 
    ALTER COLUMN owner_id SET NOT NULL;

ALTER TABLE students DROP CONSTRAINT IF EXISTS students_owner_id_fkey;
ALTER TABLE students
    ADD CONSTRAINT students_owner_id_fkey
    FOREIGN KEY (owner_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS students_owner_id_idx ON students (owner_id);