/**
 * Math utilities with intentional bugs for testing
 */

/**
 * Calculate average of an array
 * BUG: Does not handle empty arrays (division by zero)
 */
function calculateAverage(numbers) {
  const sum = numbers.reduce((acc, num) => acc + num, 0);
  // BUG: Will divide by zero if numbers is empty
  return sum / numbers.length;
}

/**
 * Calculate percentage
 * BUG: Does not handle zero denominator
 */
function calculatePercentage(part, total) {
  // BUG: Will divide by zero if total is 0
  return (part / total) * 100;
}

/**
 * Safe division (example of correct implementation)
 */
function safeDivide(numerator, denominator) {
  if (denominator === 0) {
    throw new Error('Division by zero');
  }
  return numerator / denominator;
}

module.exports = {
  calculateAverage,
  calculatePercentage,
  safeDivide
};
