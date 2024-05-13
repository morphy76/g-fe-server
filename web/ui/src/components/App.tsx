import React, { useEffect, Suspense, lazy, useState, useMemo } from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { LoggerContextProvider, useLogger } from '@features/react-logger';
import styles from '@components/App.scss';
import { useGetExample } from './examples/ExampleService';
import { BrowserRouter, Navigate, Route, Routes, useNavigate } from 'react-router-dom';

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

  const [ selected, setSelected ] = useState<string | null>(null);
  const { data, isLoading, error, refetch } = useGetExample(selected);
  const navigate = useNavigate();

  const handleExampleSelected = (name: string | null) => {
    setSelected(() => name);
  };

  const editForm = useMemo(() => (
    <>
      {isLoading && <div>Loading...</div>}
      {error && <div>Error: {error.message}</div>}
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
  ), [isLoading, error, selected, data, refetch]);


  return (
    <>
      <nav className={styles.navigation_wrapper}>
        <ul>
          <li><a onClick={() => navigate('/example', { relative: 'path' })}>Example</a></li>
          <li><a onClick={() => navigate('/credits', { relative: 'path' })}>Credits</a></li>
        </ul>
      </nav>
      <section className={styles.navigation_content}>
        <Routes>
          <Route index element={<Navigate to="/example" replace />} />
          <Route path="/example" element={
            <>
              <Suspense fallback={<div>Loading...</div>}>
                <ListExample onSelect={handleExampleSelected} />
              </Suspense>
              <Suspense fallback={<div>Loading...</div>}>
                <FormExample />
              </Suspense>
              <Suspense fallback={<div>Loading...</div>}>
                {editForm}
              </Suspense>
            </>
          } />
          <Route path="/credits" element={
            <>
              <p>Hello credits</p>
            </>
          } />
        </Routes>

      </section>
    </>
  );
};


export default App;
