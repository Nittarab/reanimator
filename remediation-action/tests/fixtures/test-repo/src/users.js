/**
 * User service with intentional bugs for testing
 */

const users = [
  { id: 1, name: 'Alice', email: 'alice@example.com' },
  { id: 2, name: 'Bob', email: 'bob@example.com' },
  { id: 3, name: 'Charlie', email: 'charlie@example.com' }
];

/**
 * Get user by ID
 * BUG: Does not handle case when user is not found
 */
function getUserById(id) {
  const user = users.find(u => u.id === id);
  // BUG: This will throw if user is undefined
  return user.id;
}

/**
 * Get user email
 * BUG: Does not validate input
 */
function getUserEmail(userId) {
  const user = users.find(u => u.id === userId);
  return user.email; // BUG: Will fail if user is undefined
}

module.exports = {
  getUserById,
  getUserEmail
};
