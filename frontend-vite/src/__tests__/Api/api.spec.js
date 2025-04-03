import { jest } from '@jest/globals';
import API from '../../api/API';

describe('API methods using URLSearchParams', () => {
  const BASE_URL = '/api';
  const applicationID = 'testAppID';
  const groupID = 'testGroupID';
  const mockFetch = jest.spyOn(global, 'fetch');

  beforeEach(() => {
    mockFetch.mockClear();
  });

  describe('getPackages', () => {
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
        description: 'a shorter search term with dot',
        searchTerm: '1.',
        queryOptions: undefined,
        expectedURL: `${BASE_URL}/apps/${applicationID}/packages?searchVersion=1.`,
      },
      {
        description: 'a search term of one digit',
        searchTerm: '1',
        queryOptions: undefined,
        expectedURL: `${BASE_URL}/apps/${applicationID}/packages?searchVersion=1`,
      },
      {
        description: 'an empty searchTerm',
        searchTerm: ' ',
        queryOptions: undefined,
        expectedURL: `${BASE_URL}/apps/${applicationID}/packages?searchVersion=+`,
      },
      {
        description: 'plus sign',
        searchTerm: '+',
        queryOptions: undefined,
        expectedURL: `${BASE_URL}/apps/${applicationID}/packages?searchVersion=%2B`,
      },
      {
        description: 'with empty string as param',
        searchTerm: undefined,
        queryOptions: { sort: '', limit: 10 },
        expectedURL: `${BASE_URL}/apps/${applicationID}/packages?limit=10`,
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

  describe('getInstances', () => {
    const testCases = [
      {
        description: 'no query parameters',
        groupID: groupID,
        queryOptions: undefined,
        expectedURL: `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instances`,
      },
      {
        description: 'query options',
        groupID: groupID,
        queryOptions: { sort: 'desc', page: 2 },
        expectedURL: `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instances?sort=desc&page=2`,
      },
      {
        description: 'empty string values omited',
        groupID: groupID,
        queryOptions: {
          status: 0,
          version: '',
          sort: 2,
          sortOrder: 0,
          page: 1,
          perpage: 10,
          duration: '1d',
        },
        expectedURL: `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instances?status=0&sort=2&sortOrder=0&page=1&perpage=10&duration=1d`,
      },
      {
        description: 'undefined values omited',
        groupID: groupID,
        queryOptions: {
          status: 0,
          version: undefined,
          sort: 2,
          sortOrder: 0,
          page: 1,
          perpage: 10,
          duration: '1d',
        },
        expectedURL: `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instances?status=0&sort=2&sortOrder=0&page=1&perpage=10&duration=1d`,
      },
      {
        description: 'null values omited',
        groupID: groupID,
        queryOptions: {
          status: 0,
          version: null,
          sort: 2,
          sortOrder: 0,
          page: 1,
          perpage: 10,
          duration: '1d',
        },
        expectedURL: `${BASE_URL}/apps/${applicationID}/groups/${groupID}/instances?status=0&sort=2&sortOrder=0&page=1&perpage=10&duration=1d`,
      },
    ];

    it.each(testCases)(
      'should construct the correct URL with $description',
      async ({ groupID, queryOptions, expectedURL }) => {
        const mockFetch = jest.fn().mockResolvedValueOnce({
          ok: true,
          json: async () => ({}),
          headers: {
            get: jest.fn().mockReturnValue('mocked-id-token'), // Mocked 'get' method
          },
        });
        global.fetch = mockFetch;

        await API.getInstances(applicationID, groupID, queryOptions);

        expect(mockFetch).toHaveBeenCalledWith(expectedURL, expect.any(Object));
      }
    );
  });
});
