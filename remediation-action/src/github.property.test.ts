/**
 * Property-based tests for GitHub operations
 * Feature: ai-sre-platform, Property 9: Branch naming includes incident ID
 * Validates: Requirements 8.1
 */

import * as fc from 'fast-check';
import { createBranch } from './github';
import * as exec from '@actions/exec';

// Mock the exec module
jest.mock('@actions/exec');

describe('GitHub operations property tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Property 9: Branch naming includes incident ID', () => {
    it('should create branch names that include the incident ID for any valid incident ID', async () => {
      // Property: For any valid incident ID, the created branch name should contain that incident ID
      
      await fc.assert(
        fc.asyncProperty(
          // Generate random incident IDs (alphanumeric with hyphens and underscores)
          fc.stringMatching(/^[a-zA-Z0-9_-]{1,50}$/),
          async (incidentId) => {
            // Mock git checkout to succeed
            const mockExec = exec.exec as jest.MockedFunction<typeof exec.exec>;
            mockExec.mockResolvedValue(0);
            
            // Create branch
            const branchName = await createBranch(incidentId);
            
            // Property: Branch name must contain the incident ID
            expect(branchName).toContain(incidentId);
            
            // Additional invariant: Branch name should follow git branch naming conventions
            // (no spaces, starts with fix/)
            expect(branchName).toMatch(/^fix\//);
            expect(branchName).not.toContain(' ');
            
            // Verify git checkout was called with the correct branch name
            expect(mockExec).toHaveBeenCalledWith('git', ['checkout', '-b', branchName]);
          }
        ),
        { numRuns: 100 }
      );
    });

    it('should create unique branch names for different incident IDs', async () => {
      // Property: Different incident IDs should produce different branch names
      
      await fc.assert(
        fc.asyncProperty(
          fc.tuple(
            fc.stringMatching(/^[a-zA-Z0-9_-]{1,50}$/),
            fc.stringMatching(/^[a-zA-Z0-9_-]{1,50}$/)
          ).filter(([id1, id2]) => id1 !== id2), // Ensure IDs are different
          async ([incidentId1, incidentId2]) => {
            const mockExec = exec.exec as jest.MockedFunction<typeof exec.exec>;
            mockExec.mockResolvedValue(0);
            
            const branchName1 = await createBranch(incidentId1);
            const branchName2 = await createBranch(incidentId2);
            
            // Property: Different incident IDs must produce different branch names
            expect(branchName1).not.toBe(branchName2);
          }
        ),
        { numRuns: 100 }
      );
    });
  });
});
