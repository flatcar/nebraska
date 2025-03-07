import { render } from '@testing-library/react';
import ListItem from '../../components/common/ListItem';

describe('List Item', () => {
  it('should render correct list item', () => {
    const { getByTestId } = render(<ListItem />);
    expect(getByTestId('list-item')).toBeInTheDocument();
  });
});
