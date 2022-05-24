// jest-dom adds custom jest matchers for asserting on DOM nodes.
// allows you to do things like:
// expect(element).toHaveTextContent(/react/i)
// learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom/extend-expect';

// Material UI doesn't have a stable ID generator.
// Every render a different ID is made and snapshot tests are broken.
// mui v5
// jest.mock('@material-ui/utils/useId', () => jest.fn().mockReturnValue('mui-test-id'))

// mui v4
jest.mock('@material-ui/core/utils/unstable_useId', () => jest.fn().mockReturnValue('mui-test-id'));
