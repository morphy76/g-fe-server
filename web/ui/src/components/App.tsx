import React, { useEffect, Suspense, lazy, useState, useMemo } from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { LoggerContextProvider, useLogger } from '@features/react-logger';
import * as styles from '@components/App.scss';
import { useGetExample } from './examples/ExampleService';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import formatted_message_rules from "@features/formatted_message_rules";
import { FormattedMessage } from 'react-intl';
import AppMenu from './AppMenu';
import { UserInfoContextProvider, useUserInfoQuery } from '@features/user-info';

const ListExample = lazy(() => import('@components/examples/ListExampleComponent'));
const FormExample = lazy(() => import('@components/examples/FormExampleComponent'));

const App: React.FC = () => {

  const logger = useLogger();

  useEffect(() => {
    logger?.log('App started');
    return () => {
      logger?.log('App stopped');
    };
  }, [logger]);

  const uiPath = useMemo(() => (
    '/' + window.location.pathname.split('/').slice(1, 3).join('/')
  ), []);

  return (
    <div className={styles.responsive_container}>
      <div className={styles.container}>
        <LoggerContextProvider component='container'>
          <QueryClientProvider client={new QueryClient()}>
            <BrowserRouter basename={uiPath}>
              <InnerApp />
            </BrowserRouter>
          </QueryClientProvider>
        </LoggerContextProvider>
      </div>
    </div>
  );
};

const InnerApp: React.FC = () => {

  const [selected, setSelected] = useState<string | null>(null);
  const { data, isLoading, error, refetch } = useGetExample(selected);
  const userInfoModel = useUserInfoQuery();

  const handleExampleSelected = (name: string | null) => {
    setSelected(() => name);
  };

  const loadingLabel = useMemo(() => (
    <FormattedMessage
      id='app.loading'
      defaultMessage='Loading...'
      values={{
        ...formatted_message_rules,
      }}
    />
  ), []);

  const errorLabel = useMemo(() => (
    <FormattedMessage
      id='app.error'
      defaultMessage='Error: {message}'
      values={{
        ...formatted_message_rules,
        message: error?.message,
      }}
    />
  ), [error]);

  const editForm = useMemo(() => (
    <>
      {isLoading && <div>{loadingLabel}</div>}
      {error && <div>{errorLabel}</div>}
      {selected && data &&
        <FormExample
          key={selected}
          example={data}
          onUpdate={() => {
            refetch();
            return true;
          }}
          onUnset={() => setSelected(null)}
        />
      }
    </>
  ), [loadingLabel, errorLabel, isLoading, error, selected, data, refetch]);

  return (
    <UserInfoContextProvider userInfo={userInfoModel.data}>
      <AppMenu />
      <section className={styles.navigation_content}>
        <Routes>
          <Route index element={<Navigate to="/example" replace />} />
          <Route path="/example" element={
            <>
              <Suspense fallback={<div>{loadingLabel}</div>}>
                <ListExample onSelect={handleExampleSelected} />
              </Suspense>
              <Suspense fallback={<div>{loadingLabel}</div>}>
                <FormExample />
              </Suspense>
              <Suspense fallback={<div>{loadingLabel}</div>}>
                {editForm}
              </Suspense>
            </>
          } />
          <Route path="/credits" element={
            <>
              <p>
                <FormattedMessage
                  id='credits.title'
                  defaultMessage='Example'
                  values={{
                    ...formatted_message_rules,
                  }}
                />
              </p>
            </>
          } />
          <Route path="*" element={
            <p>
              <FormattedMessage
                id='app.path.not.found'
                defaultMessage='Path not found'
                values={{
                  ...formatted_message_rules,
                }}
              />
            </p>
          } />
        </Routes>
      </section>
    </UserInfoContextProvider>
  );
};


export default App;
