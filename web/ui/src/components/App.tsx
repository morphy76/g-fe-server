import React, { useEffect, Suspense, lazy, useState, useMemo } from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { LoggerContextProvider, useLogger } from '@features/react-logger';
import styles from '@components/App.scss';
import { useGetExample } from './examples/ExampleService';

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

  return (
    <div className={styles.responsive_container}>
      <div className={styles.container}>
        <LoggerContextProvider component='container'>
          <QueryClientProvider client={new QueryClient()}>
            <InnerApp />
          </QueryClientProvider>
        </LoggerContextProvider>
      </div>
    </div>
  );
};

const InnerApp: React.FC = () => {

  const [ selected, setSelected ] = useState<string | null>(null);
  const { data, isLoading, error, refetch } = useGetExample(selected);

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
  );
};


export default App;
