import React from 'react';
import { createRoot } from 'react-dom/client';
import App from '@components/App';
import { loggerFor } from '@features/react-logger';
import { IntlProvider } from 'react-intl';

import English from './languages/en-US.bundle.json';
import Spanish from './languages/es-ES.bundle.json';
import French from './languages/fr-FR.bundle.json';
import Italian from './languages/it-IT.bundle.json';
import German from './languages/de-DE.bundle.json';

const locale = navigator.language;
const languageByLocale: (locale: string) => Record<string, string> = (locale) => {
  switch (locale) {
    case 'es-ES':
      return Spanish;
    case 'fr-FR':
      return French;
    case 'it-IT':
      return Italian;
    case 'de-DE':
      return German;
    default:
      return English;
  }
};

const logger = loggerFor('index');
const selectedLanguage = languageByLocale(locale);

const rootElement = document.getElementById('app');
if (rootElement) {
  const root = createRoot(rootElement);
  root.render(
    <IntlProvider locale={locale} messages={selectedLanguage}>
      <App />
    </IntlProvider>
  );
  logger.log('App started');
} else {
  logger.error('Root element not found');
}
