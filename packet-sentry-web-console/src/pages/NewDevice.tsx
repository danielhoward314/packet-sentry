import { DeviceOnboarding } from "@/components/DeviceOnboarding";
import MainContentCardLayout from "@/layouts/MainContentCardLayout";

export default function NewDevicePage() {
  return (
    <MainContentCardLayout
      cardDescription="Install the Packet Sentry Agent on a new device for remote network telemetry."
      cardTitle="New Device"
    >
      <DeviceOnboarding />
    </MainContentCardLayout>
  );
}
