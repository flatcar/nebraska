import { render } from '@testing-library/react';
import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import {
  CardDescriptionLabel,
  CardFeatureLabel,
  CardHeader,
  CardLabel,
} from '../../components/common/Card';

describe('Card', () => {
  const minProps = {
    children: 'Children Content',
    cardMainLinkPath: '/apps/123',
    cardMainLinkLabel: 'Flatcar',
  };
  it('should render correct card feature label', () => {
    const { getByText } = render(<CardFeatureLabel children={minProps.children} />);
    expect(getByText(minProps.children)).toBeInTheDocument();
  });
  it('should render correct card description label', () => {
    const { getByText } = render(<CardDescriptionLabel children={minProps.children} />);
    expect(getByText(minProps.children)).toBeInTheDocument();
  });
  it('should render correct card label', () => {
    const { getByText } = render(<CardLabel children={minProps.children} />);
    expect(getByText(minProps.children)).toBeInTheDocument();
  });
  it('should render card header with correct links', () => {
    const { container } = render(
      <BrowserRouter>
        <CardHeader
          cardMainLinkPath={minProps.cardMainLinkPath}
          cardMainLinkLabel={minProps.cardMainLinkLabel}
        />
      </BrowserRouter>
    );
    expect(container.querySelector('a').getAttribute('href')).toBe(minProps.cardMainLinkPath);
  });
  it('should render correct card label', () => {
    const { getByText } = render(<CardHeader cardMainLinkLabel={minProps.cardMainLinkLabel} />);
    expect(getByText(minProps.cardMainLinkLabel)).toBeInTheDocument();
  });
  it('should render correct card childrens', () => {
    const { getByText } = render(<CardHeader>{minProps.children}</CardHeader>);
    expect(getByText(minProps.children)).toBeInTheDocument();
  });
});
