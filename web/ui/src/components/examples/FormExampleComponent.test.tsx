import React from 'react';
import { render } from '@testing-library/react';
import FormExample from './FormExampleComponent';
import { IntlProvider } from 'react-intl';
import { LoggerContextProvider } from '@features/react-logger';
import { QueryClient, QueryClientProvider } from 'react-query';
import MockAdapter from 'axios-mock-adapter';
import { ps_client } from '@features/axios';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

describe('FormExample component', () => {

  let mock: MockAdapter | undefined;

  beforeAll(async () => {
    mock = new MockAdapter(ps_client, {
      delayResponse: 100,
    });
  });

  afterEach(async () => {
    mock?.reset();
  });

  it('should not crash while rendering when opened to create a new example', async () => {
    const component = render(
      <IntlProvider locale='en' messages={{}}>
        <LoggerContextProvider component='test'>
          <QueryClientProvider client={queryClient}>
            <FormExample />
          </QueryClientProvider>
        </LoggerContextProvider>
      </IntlProvider>
    );

    expect(component).toBeDefined();
  });

});
