import {
  Cormorant_Garamond,
  Instrument_Sans,
  JetBrains_Mono,
} from "next/font/google";

/** Display face — used at >=28px, weight <=500 only. */
export const display = Cormorant_Garamond({
  subsets: ["latin"],
  weight: ["300", "400", "500"],
  style: ["normal", "italic"],
  variable: "--font-display",
  display: "swap",
});

/** Body face — 600 carries the uppercase micro-label. */
export const instrument = Instrument_Sans({
  subsets: ["latin"],
  weight: ["400", "500", "600"],
  variable: "--font-body",
  display: "swap",
});

/** Data face — codes, prices, IDs, tabular numerics. */
export const mono = JetBrains_Mono({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-mono",
  display: "swap",
});
