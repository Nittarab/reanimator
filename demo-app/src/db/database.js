import Database from 'better-sqlite3';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Create in-memory database
const db = new Database(':memory:');

// Initialize schema
db.exec(`
  CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    age INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    price REAL NOT NULL,
    stock INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    total REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
  );
`);

// Seed some initial data
const seedData = () => {
  const insertUser = db.prepare('INSERT INTO users (name, email, age) VALUES (?, ?, ?)');
  const insertProduct = db.prepare('INSERT INTO products (name, price, stock) VALUES (?, ?, ?)');

  insertUser.run('Alice Johnson', 'alice@example.com', 30);
  insertUser.run('Bob Smith', 'bob@example.com', 25);
  insertUser.run('Charlie Brown', 'charlie@example.com', 35);

  insertProduct.run('Laptop', 999.99, 10);
  insertProduct.run('Mouse', 29.99, 50);
  insertProduct.run('Keyboard', 79.99, 30);
  insertProduct.run('Monitor', 299.99, 15);
};

seedData();

export default db;
