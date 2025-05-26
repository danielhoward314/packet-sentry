import DateTimeRangePicker from "@/components/DatetimeRangePicker";
import { EventsTable } from "@/components/EventsTable";
import MainContentCardLayout from "@/layouts/MainContentCardLayout";
import { useState } from "react";
import { useParams } from "react-router-dom";

export interface DatetimeRangeQuery {
  start: Date | null;
  end: Date | null;
}

export default function DeviceEventsPage() {
  const { deviceId } = useParams();
  const [datetimeQuery, setDatetimeQuery] = useState<DatetimeRangeQuery>({
    start: null,
    end: null,
  });
  return (
    <MainContentCardLayout
      cardDescription="Packet capture events for the selected device."
      cardTitle="Device Events"
    >
      <DateTimeRangePicker
        onChange={function (range: {
          start: Date | null;
          end: Date | null;
        }): void {
          setDatetimeQuery(range);
        }}
      />
      <EventsTable deviceId={deviceId ?? ""} datetimeQuery={datetimeQuery} />
    </MainContentCardLayout>
  );
}
