import { Navigate, Outlet, Route, Routes } from "react-router";

import { TooltipProvider } from "./components/ui/tooltip";
import { Knowing } from "./lib/Here";
import { Toaster } from "./components/ui/toast";
import { NotFound } from "./routes/NotFound";
import { OrganisationPage } from "./routes/OrganisationPage";
import { OnboardingPage } from "./routes/OnboardingPage";
import { HistoryPage } from "./routes/HistoryPage";
import { ContentPage } from "./routes/ContentPage";
import { EditorPage } from "./routes/EditorPage";
import { MediaPage } from "./routes/MediaPage";
import { PeoplePage } from "./routes/PeoplePage";
import { SyncPage } from "./routes/SyncPage";
import { PublishPage } from "./routes/PublishPage";
import { ProjectPage } from "./routes/ProjectPage";
import { AccountPage } from "./routes/AccountPage";
import { AdministrationPage } from "./routes/AdministrationPage";
import { OrganisationSettingsPage } from "./routes/OrganisationSettingsPage";
import { ProjectSettingsPage } from "./routes/ProjectSettingsPage";
import { ProjectsPage } from "./routes/ProjectsPage";
import { ConfirmAddressPage } from "./routes/ConfirmAddressPage";
import { ForgotPasswordPage } from "./routes/ForgotPasswordPage";
import { SetPasswordPage } from "./routes/SetPasswordPage";
import { SignInPage } from "./routes/SignInPage";

export function App() {
  return (
    <TooltipProvider delayDuration={200}>
      <Routes>
        <Route path="/" element={<Navigate to="/projects" replace />} />
        <Route path="/sign-in" element={<SignInPage />} />
        <Route path="/set-password" element={<SetPasswordPage />} />
        <Route path="/confirm-address" element={<ConfirmAddressPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />

        {/* Everything below knows who is looking and which organisation the
            address names, asked once rather than guessed per screen. */}
        <Route element={<Knowing><Outlet /></Knowing>}>
          <Route path="/projects" element={<ProjectsPage />} />
          <Route path="/help" element={<OrganisationPage title="Help" />} />
          <Route
            path="/account"
            element={<OrganisationPage title="Your account" body={<AccountPage />} />}
          />
          <Route
            path="/administration"
            element={<OrganisationPage title="Administration" body={<AdministrationPage />} />}
          />
        </Route>

        <Route path="/o/:organisation" element={<Knowing><Outlet /></Knowing>}>
          <Route path="projects" element={<ProjectsPage />} />
          <Route
            path="settings"
            element={<OrganisationPage title="Organisation" body={<OrganisationSettingsPage />} />}
          />

          <Route path="p/:project/setup" element={<ProjectPage title="Set up a project" body={<OnboardingPage />} />} />
          <Route path="p/:project/people" element={<ProjectPage title="People & roles" body={<PeoplePage />} />} />
          <Route
            path="p/:project/settings"
            element={<ProjectPage title="Project settings" body={<ProjectSettingsPage />} />}
          />
          <Route path="p/:project/workflows" element={<ProjectPage title="Workflows" />} />
          <Route path="p/:project/content" element={<ProjectPage title="Content" body={<ContentPage />} />} />
          <Route path="p/:project/content/*" element={<EditorPage />} />
          <Route path="p/:project/media" element={<ProjectPage title="Media" body={<MediaPage />} />} />
          <Route path="p/:project/publish" element={<ProjectPage title="Publish" body={<PublishPage />} />} />
          <Route path="p/:project/history" element={<ProjectPage title="History" body={<HistoryPage />} />} />
          <Route path="p/:project/sync" element={<ProjectPage title="Sync status" body={<SyncPage />} />} />
        </Route>

        <Route path="*" element={<NotFound />} />
      </Routes>
      <Toaster />
    </TooltipProvider>
  );
}
