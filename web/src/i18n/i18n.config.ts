import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import en from './locales/en.json';

const resources = {
  en: {
    translation: en,
  },
};

i18n
  .use(initReactI18next) // Pass i18n down to react-i18next
  .init({
    resources,
    lng: 'en', // English only
    fallbackLng: 'en',
    supportedLngs: ['en'], // Only English supported

    interpolation: {
      escapeValue: false, // React already escapes
    },

    react: {
      useSuspense: false, // Disable suspense to avoid loading issues
    },
  });

// Set initial language to English and LTR
document.documentElement.lang = 'en';
document.documentElement.dir = 'ltr';

export default i18n;

// Language configuration (English only)
export const languages = [
  {
    code: 'en',
    name: 'English',
    nativeName: 'English',
    dir: 'ltr',
    flag: '🇺🇸',
  },
];

// Get language configuration
export const getLanguageConfig = (code: string) => {
  return languages.find((lang) => lang.code === code) || languages[0];
};

// Change language
export const changeLanguage = (code: string) => {
  const config = getLanguageConfig(code);
  if (config) {
    i18n.changeLanguage(code);
    document.documentElement.dir = config.dir;
    document.documentElement.lang = code;
    localStorage.setItem('edgeflow-language', code);
  }
};

// Get current language
export const getCurrentLanguage = () => {
  return i18n.language || 'en';
};

// Get current direction
export const getCurrentDirection = () => {
  const config = getLanguageConfig(getCurrentLanguage());
  return config.dir;
};
