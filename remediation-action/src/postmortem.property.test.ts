/**
 * Property-based tests for post-mortem generation
 * Feature: ai-sre-platform, Property 10: Post-mortem completeness
 * Validates: Requirements 8.4, 8.5
 */

import * as fc from 'fast-check';
import { generatePostMortem, validatePostMortem, PostMortemData } from './postmortem';

describe('Post-mortem generation property tests', () => {
  describe('Property 10: Post-mortem completeness', () => {
    it('should generate post-mortems with all required sections for any valid input', async () => {
      // Property: For any valid post-mortem data, the generated post-mortem should contain
      // all required sections: what broke, why it broke, how the fix addresses it, and test results
      
      await fc.assert(
        fc.asyncProperty(
          // Generate random post-mortem data
          fc.record({
            incidentId: fc.stringMatching(/^[a-zA-Z0-9_-]{1,50}$/),
            serviceName: fc.string({ minLength: 1, maxLength: 100 }),
            timestamp: fc.date().map(d => d.toISOString()),
            errorMessage: fc.string({ minLength: 1, maxLength: 500 }),
            stackTrace: fc.option(fc.string({ minLength: 1, maxLength: 1000 }), { nil: undefined }),
            diagnosis: fc.string({ minLength: 1, maxLength: 1000 }),
            fixDescription: fc.string({ minLength: 1, maxLength: 1000 }),
            testResults: fc.option(fc.string({ minLength: 1, maxLength: 500 }), { nil: undefined }),
          }),
          async (data: PostMortemData) => {
            // Generate post-mortem
            const postMortem = generatePostMortem(data);
            
            // Property: Post-mortem must contain all required sections
            expect(postMortem).toContain('What Broke');
            expect(postMortem).toContain('Why It Broke');
            expect(postMortem).toContain('How the Fix Addresses It');
            expect(postMortem).toContain('Test Results');
            
            // Property: Post-mortem must include the incident ID
            expect(postMortem).toContain(data.incidentId);
            
            // Property: Post-mortem must include the service name
            expect(postMortem).toContain(data.serviceName);
            
            // Property: Post-mortem must include the error message
            expect(postMortem).toContain(data.errorMessage);
            
            // Property: Post-mortem must include the diagnosis
            expect(postMortem).toContain(data.diagnosis);
            
            // Property: Post-mortem must include the fix description
            expect(postMortem).toContain(data.fixDescription);
            
            // Property: If stack trace is provided, it must be included
            if (data.stackTrace) {
              expect(postMortem).toContain(data.stackTrace);
            }
            
            // Property: If test results are provided, they must be included
            if (data.testResults) {
              expect(postMortem).toContain(data.testResults);
            }
            
            // Property: Validation should pass for generated post-mortems
            expect(validatePostMortem(postMortem)).toBe(true);
          }
        ),
        { numRuns: 100 }
      );
    });

    it('should validate that post-mortems contain all required sections', async () => {
      // Property: For any post-mortem, validation should correctly identify missing sections
      
      await fc.assert(
        fc.asyncProperty(
          fc.record({
            incidentId: fc.stringMatching(/^[a-zA-Z0-9_-]{1,50}$/),
            serviceName: fc.string({ minLength: 1, maxLength: 100 }),
            timestamp: fc.date().map(d => d.toISOString()),
            errorMessage: fc.string({ minLength: 1, maxLength: 500 }),
            diagnosis: fc.string({ minLength: 1, maxLength: 1000 }),
            fixDescription: fc.string({ minLength: 1, maxLength: 1000 }),
          }),
          async (data) => {
            // Generate a complete post-mortem
            const completePostMortem = generatePostMortem({
              ...data,
              testResults: 'All tests passed',
            });
            
            // Property: Complete post-mortem should pass validation
            expect(validatePostMortem(completePostMortem)).toBe(true);
            
            // Property: Removing any required section should fail validation
            const sections = ['What Broke', 'Why It Broke', 'How the Fix Addresses It', 'Test Results'];
            
            for (const section of sections) {
              const incompletePostMortem = completePostMortem.replace(section, '');
              expect(validatePostMortem(incompletePostMortem)).toBe(false);
            }
          }
        ),
        { numRuns: 100 }
      );
    });

    it('should generate consistent post-mortems for the same input', async () => {
      // Property: Generating a post-mortem twice with the same input should produce identical results
      
      await fc.assert(
        fc.asyncProperty(
          fc.record({
            incidentId: fc.stringMatching(/^[a-zA-Z0-9_-]{1,50}$/),
            serviceName: fc.string({ minLength: 1, maxLength: 100 }),
            timestamp: fc.date().map(d => d.toISOString()),
            errorMessage: fc.string({ minLength: 1, maxLength: 500 }),
            diagnosis: fc.string({ minLength: 1, maxLength: 1000 }),
            fixDescription: fc.string({ minLength: 1, maxLength: 1000 }),
            testResults: fc.option(fc.string({ minLength: 1, maxLength: 500 }), { nil: undefined }),
          }),
          async (data: PostMortemData) => {
            const postMortem1 = generatePostMortem(data);
            const postMortem2 = generatePostMortem(data);
            
            // Property: Same input should produce identical output (deterministic)
            expect(postMortem1).toBe(postMortem2);
          }
        ),
        { numRuns: 100 }
      );
    });
  });
});
