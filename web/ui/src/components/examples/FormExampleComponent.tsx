import React, { useCallback, useEffect, useId, useMemo } from "react";
import styles from './FormExampleComponent.scss';
import { Example, useCreateExample, useReplaceExample } from "./ExampleService";
import { FormattedMessage } from "react-intl";
import formatted_message_rules from "@features/formatted_message_rules";
import { useLogger } from "@features/react-logger";
import { useForm, SubmitHandler } from "react-hook-form";

type FormExampleProps = {
  example?: Example;
  onCreate?: (example: Example) => boolean;
  onUpdate?: (example: Example) => void;
  onReset?: () => void;
  onUnset?: () => void;
};

const FormExample: React.FC<FormExampleProps> = ({ example, onCreate, onUpdate, onReset, onUnset }) => {

  const { register, handleSubmit, reset, formState: { isDirty, isSubmitting, errors, isSubmitSuccessful } } = useForm<Example>();
  const createQuery = useCreateExample();
  const updateQuery = useReplaceExample(example?.name ?? '');
  const logger = useLogger();

  const initialValue = useMemo(() => example ?? {
    name: '',
    age: 0,
  }, [example]);

  useEffect(() => {
    logger?.debug('FormExample mounted');
    return () => {
      logger?.debug('FormExample unmounted');
    };
  }, [logger]);

  useEffect(() => {
    reset(initialValue);
  }, [initialValue, reset]);

  const onFormUnset = useCallback(() => {
    if (onUnset) {
      onUnset();
    }
  }, [onUnset]);

  const onFormReset = useCallback(() => {
    reset(initialValue);
    if (onReset) {
      onReset();
    }
  }, [reset, initialValue, onReset]);

  const onFormSubmit: SubmitHandler<Example> = useCallback(async (formExample: Example) => {
    logger.debug('FormExample submit', formExample);
    if (example) {
      await updateQuery.mutateAsync(formExample);
      if (onUpdate) {
        onUpdate(formExample);
      }
    } else {
      await createQuery.mutateAsync(formExample);
      if (onCreate) {
        onCreate(formExample) && reset(initialValue);
      } else {
        reset(initialValue);
      }
    }
  }, [logger, example, onUpdate, onCreate, reset, initialValue, createQuery, updateQuery]);

  const idName = useId();
  const idAge = useId();
  return (
    <form className={styles.form_example_wrapper} onSubmit={handleSubmit(onFormSubmit)}>
      <fieldset className={styles.responsive_fieldset}>
        <legend>
          {example && <FormattedMessage
            id='examples.form.legend.existing'
            defaultMessage='Edit Example'
            values={{
              ...formatted_message_rules,
            }}
          />}
          {!example && <FormattedMessage
            id='examples.form.legend.new'
            defaultMessage='New Example'
            values={{
              ...formatted_message_rules,
            }}
          />}
        </legend>
        <div className={styles.responsive_field}>
          <label htmlFor={idName}>
            <FormattedMessage
              id='examples.form.name'
              defaultMessage='Name'
              values={{
                ...formatted_message_rules,
              }}
            />
          </label>
          <input
            type="text"
            id={idName}
            tabIndex={0}
            {...register('name', { required: true })}
            aria-invalid={errors.name ? "true" : "false"}
            readOnly={!!example}
          />
        </div>
        <div className={styles.responsive_field}>
          <label htmlFor={idAge}>
            <FormattedMessage
              id='examples.form.age'
              defaultMessage='Age'
              values={{
                ...formatted_message_rules,
              }}
            />
          </label>
          <input
            type="number"
            id={idAge}
            tabIndex={0}
            {...register('age', { min: 1, max: 200, required: true })}
            aria-invalid={errors.age ? "true" : "false"}
          />
        </div>
        <div className={styles.responsive_field}>
          {example && <button
            onClick={onFormUnset}
          >
            <FormattedMessage
              id='examples.form.close'
              defaultMessage='Close'
              values={{
                ...formatted_message_rules,
              }}
            />
          </button>}
          <button
            tabIndex={0}
            onClick={onFormReset}
            disabled={isSubmitting || !isDirty}
          >
            {example && <FormattedMessage
              id='examples.form.reset.existing'
              defaultMessage='Discard'
              values={{
                ...formatted_message_rules,
              }}
            />}
            {!example && <FormattedMessage
              id='examples.form.reset.new'
              defaultMessage='Clear'
              values={{
                ...formatted_message_rules,
              }}
            />}
          </button>
          <button
            type="submit"
            tabIndex={0}
            disabled={isSubmitting || !isDirty}
          >
            {example && <FormattedMessage
              id='examples.form.submit.existing'
              defaultMessage='Update'
              values={{
                ...formatted_message_rules,
              }}
            />}
            {!example && <FormattedMessage
              id='examples.form.submit.new'
              defaultMessage='Create'
              values={{
                ...formatted_message_rules,
              }}
            />}
          </button>
        </div>
      </fieldset>
      <div>
        {isSubmitSuccessful &&
          <p className={styles.form_saved}>
            <FormattedMessage
              id='examples.form.saved'
              defaultMessage='Example saved'
              values={{
                ...formatted_message_rules,
              }}
            />
          </p>}
        {errors.name?.type === "required" &&
          <p className={styles.error_message}>
            <FormattedMessage
              id='examples.form.errors.required.name'
              defaultMessage='Name is required'
              values={{
                ...formatted_message_rules,
              }}
            />
          </p>}
        {errors.age &&
          <p className={styles.error_message}>
            <FormattedMessage
              id='examples.form.errors.invalid.age'
              defaultMessage='Age must be positive'
              values={{
                ...formatted_message_rules,
              }}
            />
          </p>}
      </div>
    </form>
  );
};

export default FormExample;
