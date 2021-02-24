import React, { Suspense } from "react";

const LightTheme = React.lazy(() => import("./LightTheme"));
const DarkTheme = React.lazy(() => import("./DarkTheme"));

interface IThemeSelectorProps {
  theme: string;
  children: React.ReactNode;
}

export enum SupportedThemes {
  dark = "dark",
  light = "light",
}

export default function ThemeSelector({ theme, children }: IThemeSelectorProps) {
  return (
    <>
      <Suspense fallback={null}>
        {theme === SupportedThemes.light && <LightTheme />}
        {theme === SupportedThemes.dark && <DarkTheme />}
      </Suspense>
      {children}
    </>
  );
}
