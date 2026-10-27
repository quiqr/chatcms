import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router";

import { App } from "./App";
import "./index.css";
import { applyWhatWasRemembered } from "./lib/appearance";

const root = document.getElementById("root");
if (!root) {
  throw new Error("The page is missing its root element.");
}

// Before the first paint, so that nobody sees the wrong appearance flash.
applyWhatWasRemembered();

createRoot(root).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
);
