import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@realtime-me/web-ui/chart";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { type ChartPoint, type ChartUnit, chartValue, formatChartTime } from "@/lib/format";

const AREA_CHART_CONFIG = {
  value: {
    label: "Value",
    color: "var(--chart-1)",
  },
} satisfies ChartConfig;

export default function StatusChart({ data, unit }: { data: ChartPoint[]; unit: ChartUnit }) {
  return (
    <ChartContainer config={AREA_CHART_CONFIG} className="h-56 w-full">
      <AreaChart data={data} margin={{ left: 0, right: 8, top: 8, bottom: 0 }}>
        <CartesianGrid vertical={false} />
        <XAxis
          dataKey="time"
          tickFormatter={formatChartTime}
          tickLine={false}
          axisLine={false}
          minTickGap={32}
        />
        <YAxis
          tickFormatter={(value) => chartValue(unit, Number(value))}
          tickLine={false}
          axisLine={false}
          width={56}
        />
        <ChartTooltip
          cursor={false}
          content={
            <ChartTooltipContent hideLabel formatter={(value) => chartValue(unit, Number(value))} />
          }
        />
        <Area
          dataKey="value"
          type="monotone"
          stroke="var(--color-value)"
          fill="var(--color-value)"
          fillOpacity={0.2}
          strokeWidth={2}
          isAnimationActive={false}
        />
      </AreaChart>
    </ChartContainer>
  );
}
