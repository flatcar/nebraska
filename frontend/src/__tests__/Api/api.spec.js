import { jest } from '@jest/globals';
import API from '../../api/API';

describe('getPackages', () => {
  const BASE_URL = '/api';
  const applicationID = 'testAppID';
  const mockFetch = jest.spyOn(global, 'fetch');

  beforeEach(() => {
    mockFetch.mockClear();
  });

  // Parameterized test cases
  const testCases = [
    {
      description: 'no query parameters',
      searchTerm: undefined,
      queryOptions: undefined,
      expectedURL: `${BASE_URL}/apps/${applicationID}/packages`,
    },
    {
      description: 'a search term',
      searchTerm: '1.2.0',
      queryOptions: undefined,
      expectedURL: `${BASE_URL}/apps/${applicationID}/packages?searchVersion=1.2.0`,
    },
    {
      description: 'a search term',
      searchTerm: '1.',
      queryOptions: undefined,
      expectedURL: `${BASE_URL}/apps/${applicationID}/packages?searchVersion=1.`,
    },
    {
      description: 'a search term',
      searchTerm: '1',
      queryOptions: undefined,
      expectedURL: `${BASE_URL}/apps/${applicationID}/packages?searchVersion=1`,
    },
    {
      description: 'query options',
      searchTerm: undefined,
      queryOptions: { sort: 'asc', limit: 10 },
      expectedURL: `${BASE_URL}/apps/${applicationID}/packages?sort=asc&limit=10`,
    },
    {
      description: 'both search term and query options',
      searchTerm: '1.0.0',
      queryOptions: { sort: 'asc', limit: 10 },
      expectedURL: `${BASE_URL}/apps/${applicationID}/packages?searchVersion=1.0.0&sort=asc&limit=10`,
    },
  ];

  it.each(testCases)(
    'should construct the correct URL with $description',
    async ({ searchTerm, queryOptions, expectedURL }) => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
        headers: {
          get: jest.fn().mockReturnValue('mocked-id-token'), // Mocked 'get' method
        },
      });

      await API.getPackages(applicationID, searchTerm, queryOptions);

      expect(mockFetch).toHaveBeenCalledWith(expectedURL, expect.any(Object));
    }
  );
});
