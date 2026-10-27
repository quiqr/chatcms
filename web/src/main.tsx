import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router";

import { App } from "./App";
import "./index.css";
import { applyWhatWasRemembered } from "./lib/appearance";
import { speakWhatWasRemembered } from "./lib/say";

const root = document.getElementById("root");
if (!root) {
  // chatcms-not-a-string: there is no interface yet to say it in.
  throw new Error("The page is missing its root element.");
}

// Before the first paint, so that nobody sees the wrong appearance flash.
applyWhatWasRemembered();
speakWhatWasRemembered();

createRoot(root).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
);
