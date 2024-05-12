import React, { act } from 'react';
import { ps_client } from "../../features/axios";
import MockAdapter from "axios-mock-adapter";
import { render } from '@testing-library/react';
import ListExample from './ListExampleComponent';
import { QueryClient, QueryClientProvider } from 'react-query';
import { LoggerContextProvider } from '../../features/react-logger';
import { IntlProvider } from 'react-intl';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

describe('ListExample component', () => {

  let mock: MockAdapter | undefined;

  beforeAll(async () => {
    mock = new MockAdapter(ps_client, {
      delayResponse: 100,
    });
  });

  afterEach(async () => {
    mock?.reset();
  });

  it('should not crash while fetching data', async () => {

    mock?.onGet('/example').reply(200, [
      { name: "John", age: 30 },
      { name: "Jane", age: 25 }
    ]);

    const component = await act(() => render(
      <IntlProvider locale='en' messages={{}}>
        <LoggerContextProvider component='test'>
          <QueryClientProvider client={queryClient}>
            <ListExample />
          </QueryClientProvider>
        </LoggerContextProvider>
      </IntlProvider>
    ));

    const loadingElement = await component.findByText((content) => content.startsWith('Loading...'));
    expect(loadingElement).toBeDefined();
  });

  it('should not crash as data is fetched', async () => {

    mock?.onGet('/example').reply(200, [
      { name: "John", age: 30 },
      { name: "Jane", age: 25 }
    ]);

    const component = await act(() => render(
      <IntlProvider locale='en' messages={{}}>
        <LoggerContextProvider component='test'>
          <QueryClientProvider client={queryClient}>
            <ListExample />
          </QueryClientProvider>
        </LoggerContextProvider>
      </IntlProvider>
    ));

    const johnElement = await component.findByText((content) => content.startsWith('John'));
    expect(johnElement).toBeDefined();

    const janeElement = await component.findByText((content) => content.startsWith('Jane'));
    expect(janeElement).toBeDefined();
  });

  it('should not crash when the request fails', async () => {

    mock?.onGet('/example').networkError();

    const component = await act(() => render(
      <IntlProvider locale='en' messages={{}}>
        <LoggerContextProvider component='test'>
          <QueryClientProvider client={queryClient}>
            <ListExample />
          </QueryClientProvider>
        </LoggerContextProvider>
      </IntlProvider>
    ));

    const errorElement = await component.findByText((content) => content.startsWith('Error:'));
    expect(errorElement).toBeDefined();
    expect(errorElement.textContent).toContain('Network Error');
  });
});
