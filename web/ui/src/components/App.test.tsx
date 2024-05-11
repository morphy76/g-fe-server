import React from 'react';
import { render } from '@testing-library/react';
import App from '../components/App';
import { IntlProvider } from 'react-intl';

describe('App component', () => {
  it('should not crash', () => {
    render(
      <IntlProvider locale='en' messages={{}}>
        <App />
      </IntlProvider>
    );
  });
});
