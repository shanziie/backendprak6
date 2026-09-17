-- Membuat tabel roles untuk manajemen hak akses berbasis peran (RBAC)
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Membuat tabel permissions untuk izin akses spesifik
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Membuat tabel relasi role_permissions (Many-to-Many)
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Menambahkan data role bawaan awal
INSERT INTO roles (name, description) 
VALUES 
    ('admin', 'Administrator dengan hak akses penuh'),
    ('user', 'Pengguna standar aplikasi')
ON CONFLICT (name) DO NOTHING;