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
import { AdministeredOrganisationPage } from "./routes/AdministeredOrganisationPage";
import { AdministeredPersonPage } from "./routes/AdministeredPersonPage";
import { AdministrationPage } from "./routes/AdministrationPage";
import { OrganisationSettingsPage } from "./routes/OrganisationSettingsPage";
import { PreferencesPage } from "./routes/PreferencesPage";
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
          <Route path="/help" element={<OrganisationPage title="common.help" />} />
          <Route
            path="/account"
            element={<OrganisationPage title="app.your-account" body={<AccountPage />} />}
          />
          <Route
            path="/preferences"
            element={<OrganisationPage title="app.preferences" body={<PreferencesPage />} />}
          />
          <Route
            path="/administration"
            element={<OrganisationPage title="app.administration" body={<AdministrationPage />} />}
          />
          <Route
            path="/administration/people/:email"
            element={<OrganisationPage title="common.person" body={<AdministeredPersonPage />} />}
          />
          <Route
            path="/administration/o/:organisation"
            element={
              <OrganisationPage title="app.organisation" body={<AdministeredOrganisationPage />} />
            }
          />
        </Route>

        <Route path="/o/:organisation" element={<Knowing><Outlet /></Knowing>}>
          <Route path="projects" element={<ProjectsPage />} />
          <Route
            path="settings"
            element={<OrganisationPage title="app.organisation" body={<OrganisationSettingsPage />} />}
          />

          <Route path="p/:project/setup" element={<ProjectPage title="app.set-up-a-project" body={<OnboardingPage />} />} />
          <Route path="p/:project/people" element={<ProjectPage title="app.people-roles" body={<PeoplePage />} />} />
          <Route
            path="p/:project/settings"
            element={<ProjectPage title="app.project-settings" body={<ProjectSettingsPage />} />}
          />
          <Route path="p/:project/workflows" element={<ProjectPage title="app.workflows" />} />
          <Route path="p/:project/content" element={<ProjectPage title="app.content" body={<ContentPage />} />} />
          <Route path="p/:project/content/*" element={<EditorPage />} />
          <Route path="p/:project/media" element={<ProjectPage title="app.media" body={<MediaPage />} />} />
          <Route path="p/:project/publish" element={<ProjectPage title="common.publish" body={<PublishPage />} />} />
          <Route path="p/:project/history" element={<ProjectPage title="app.history" body={<HistoryPage />} />} />
          <Route path="p/:project/sync" element={<ProjectPage title="app.sync-status" body={<SyncPage />} />} />
        </Route>

        <Route path="*" element={<NotFound />} />
      </Routes>
      <Toaster />
    </TooltipProvider>
  );
}
