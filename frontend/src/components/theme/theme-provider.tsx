"use client";

import { createContext, useContext, type ReactNode } from "react";

import { APP_THEME, type AppTheme } from "@/constants/themes";

const ThemeContext = createContext<AppTheme>(APP_THEME.storefront);

type ThemeProviderProps = {
  children: ReactNode;
  theme: AppTheme;
};

export function ThemeProvider({ children, theme }: ThemeProviderProps) {
  return (
    <ThemeContext.Provider value={theme}>{children}</ThemeContext.Provider>
  );
}

export function useAppTheme(): AppTheme {
  return useContext(ThemeContext);
}
