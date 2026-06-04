import React from 'react';
import ReactDOM from 'react-dom/client';
import { App } from './app/App';
import { I18nProvider } from './shared/i18n/i18n';
import { ThemeProvider } from './shared/theme/ThemeProvider';
import './styles.css';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <I18nProvider>
      <ThemeProvider>
        <App />
      </ThemeProvider>
    </I18nProvider>
  </React.StrictMode>,
);
