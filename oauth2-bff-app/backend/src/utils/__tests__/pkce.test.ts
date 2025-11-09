import { generateState, generateCodeVerifier, generateCodeChallenge, validateState } from '../pkce';

describe('PKCE Utilities', () => {
  describe('generateState', () => {
    it('should generate a random state string', () => {
      const state = generateState();
      expect(state).toBeTruthy();
      expect(typeof state).toBe('string');
      expect(state.length).toBeGreaterThan(0);
    });

    it('should generate unique states', () => {
      const state1 = generateState();
      const state2 = generateState();
      expect(state1).not.toBe(state2);
    });

    it('should generate base64url encoded string', () => {
      const state = generateState();
      // Base64url should not contain +, /, or =
      expect(state).not.toMatch(/[+/=]/);
    });
  });

  describe('generateCodeVerifier', () => {
    it('should generate a random code verifier', () => {
      const verifier = generateCodeVerifier();
      expect(verifier).toBeTruthy();
      expect(typeof verifier).toBe('string');
      expect(verifier.length).toBeGreaterThan(0);
    });

    it('should generate unique verifiers', () => {
      const verifier1 = generateCodeVerifier();
      const verifier2 = generateCodeVerifier();
      expect(verifier1).not.toBe(verifier2);
    });

    it('should generate base64url encoded string', () => {
      const verifier = generateCodeVerifier();
      expect(verifier).not.toMatch(/[+/=]/);
    });
  });

  describe('generateCodeChallenge', () => {
    it('should generate a code challenge from verifier', () => {
      const verifier = generateCodeVerifier();
      const challenge = generateCodeChallenge(verifier);
      
      expect(challenge).toBeTruthy();
      expect(typeof challenge).toBe('string');
      expect(challenge.length).toBeGreaterThan(0);
    });

    it('should generate same challenge for same verifier', () => {
      const verifier = 'test-verifier';
      const challenge1 = generateCodeChallenge(verifier);
      const challenge2 = generateCodeChallenge(verifier);
      
      expect(challenge1).toBe(challenge2);
    });

    it('should generate different challenges for different verifiers', () => {
      const verifier1 = 'test-verifier-1';
      const verifier2 = 'test-verifier-2';
      const challenge1 = generateCodeChallenge(verifier1);
      const challenge2 = generateCodeChallenge(verifier2);
      
      expect(challenge1).not.toBe(challenge2);
    });

    it('should generate base64url encoded string', () => {
      const verifier = generateCodeVerifier();
      const challenge = generateCodeChallenge(verifier);
      expect(challenge).not.toMatch(/[+/=]/);
    });
  });

  describe('validateState', () => {
    it('should return true for matching states', () => {
      const state = 'test-state-123';
      expect(validateState(state, state)).toBe(true);
    });

    it('should return false for non-matching states', () => {
      const state1 = 'test-state-123';
      const state2 = 'test-state-456';
      expect(validateState(state1, state2)).toBe(false);
    });

    it('should return false for empty states', () => {
      expect(validateState('', 'test')).toBe(false);
      expect(validateState('test', '')).toBe(false);
    });
  });
});
