import React, { PropsWithChildren, createContext, useContext } from 'react';

export type Logger = {
  error: (message?: string, ...optionalParams: unknown[]) => void;
  log: (message?: string, ...optionalParams: unknown[]) => void;
  debug: (message?: string, ...optionalParams: unknown[]) => void;
};

const LoggerContext = createContext<Logger | undefined>(undefined);

export const useLogger = (): Logger => {
  return useContext(LoggerContext)!;
};

export const loggerFor: (component: string) => Logger = (component: string) => ({
  debug: (message, ...optionalParams) => {
    if (process.env.NODE_ENV === 'development') {
      console.debug(`${new Date()} - ${component} - ${message}`, ...optionalParams);
    }
  },
  log: (message, ...optionalParams) => {
    console.log(`${new Date()} - ${component} - ${message}`, ...optionalParams);
  },
  error: (message, ...optionalParams) => {
    console.error(`${new Date()} - ${component} - ${message}`, ...optionalParams);
  },
});

type LoggerContextProviderProps = PropsWithChildren<{
  component: string
}>;
export const LoggerContextProvider: React.FC<LoggerContextProviderProps> = ({ component, children }) => (
  <LoggerContext.Provider value={loggerFor(component)}>
    {children}
  </LoggerContext.Provider>
);
