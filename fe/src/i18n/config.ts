export const locales = ['en', 'id'] as const;

export type Locale = (typeof locales)[number];


export const defaultLocale = 'id';

export const localeNames: Record<Locale, string> = {
    en: 'English',
    id: 'Bahasa Indonesia'
};

export const localeFlags: Record<Locale, string> = {
    en: '🇺🇸',
    id: '🇮🇩'
}