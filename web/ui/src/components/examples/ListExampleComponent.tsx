import React, { useEffect, useMemo } from 'react';
import { Example, useDeleteExample, useListExampleQuery } from './ExampleService';
import { useLogger } from '@features/react-logger';
import * as styles from './ListExampleComponent.scss';
import { FormattedMessage, useIntl } from 'react-intl';
import formatted_message_rules from '@features/formatted_message_rules';

type ListExampleProps = {
  onSelect?: (name: string | null) => void;
};
const ListExample: React.FC<ListExampleProps> = ({ onSelect }) => {

  const logger = useLogger();
  const { isLoading, isError, data, error } = useListExampleQuery();
  const { mutateAsync } = useDeleteExample();
  const intl = useIntl();

  useEffect(() => {
    logger?.debug('ListExample mounted');
    return () => {
      logger?.debug('ListExample unmounted');
    };
  }, [logger]);

  const content = useMemo(() => {

    const handleClick = (name: string) => {
      if (onSelect) {
        onSelect(name);
      }
    };

    const handleDelete = (name: string) => {
      const confirmMessage = intl.formatMessage({
        id: "examples.list.confirm.delete",
        defaultMessage: "Are you sure you want to delete the example [{name}]?",
      }, {
        ...formatted_message_rules,
        name: name
      }) as string;
      if (confirm(confirmMessage)) {
        mutateAsync(name);
        if (onSelect) {
          onSelect(null);
        }
      }
    };

    return data?.map((example: Example) => (
      <ListItemExample
        key={`row-${example.name}`}
        item={example}
        handleClick={handleClick}
        handleDelete={handleDelete}
      />
    ));
  }, [data, onSelect, mutateAsync, intl]);

  return (
    <div className={styles.list_example_wrapper}>
      <header>
        <FormattedMessage
          id='examples.title'
          defaultMessage='Live Examples'
          values={{
            ...formatted_message_rules,
            count: data?.length ?? 0,
          }}
        />
      </header>
      {isLoading && <div className={styles.loading}>Loading...</div>}
      {!isLoading && !isError && <ul>{content}</ul>}
      {isError && <div className={styles.error}>Error: {error.message}</div>}
    </div>
  );
};

type ListItemExampleProps = {
  item: Example;
  handleClick: (name: string) => void;
  handleDelete: (name: string) => void;
};
const ListItemExample: React.FC<ListItemExampleProps> = ({ item, handleClick, handleDelete }) => {

  const content = useMemo(() => (
    <>
      <button
        key={item.name}
        onClick={() => handleClick(item.name)}
      >
        <div>{item.name}</div>
        <div>({item.age})</div>
      </button>
      <button
        key={`del-${item.name}`}
        onClick={() => handleDelete(item.name)}
      >Del</button>
    </>
  ), [item, handleClick, handleDelete]);

  return (
    <span>{content}</span>
  );
};

export default ListExample;
