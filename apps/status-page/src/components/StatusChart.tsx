import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from 'recharts';
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from '@/components/ui/chart';

type ChartUnit = 'percent' | 'bytes' | 'count' | 'rate';

type ChartPoint = {
  time: number;
  value: number;
};

const AREA_CHART_CONFIG = {
  value: {
    label: 'Value',
    color: 'var(--chart-1)',
  },
} satisfies ChartConfig;

export default function StatusChart({ data, unit }: { data: ChartPoint[]; unit: ChartUnit }) {
  return (
    <ChartContainer config={AREA_CHART_CONFIG} className="h-56 w-full">
      <AreaChart data={data} margin={{ left: 0, right: 8, top: 8, bottom: 0 }}>
        <CartesianGrid vertical={false} />
        <XAxis dataKey="time" tickFormatter={formatChartTime} tickLine={false} axisLine={false} minTickGap={32} />
        <YAxis tickFormatter={(value) => chartValue(unit, Number(value))} tickLine={false} axisLine={false} width={56} />
        <ChartTooltip cursor={false} content={<ChartTooltipContent hideLabel formatter={(value) => chartValue(unit, Number(value))} />} />
        <Area dataKey="value" type="monotone" stroke="var(--color-value)" fill="var(--color-value)" fillOpacity={0.2} strokeWidth={2} isAnimationActive={false} />
      </AreaChart>
    </ChartContainer>
  );
}

function chartValue(unit: ChartUnit, value: number): string {
  if (unit === 'bytes') return formatGigabytes(value);
  if (unit === 'percent') return formatPercent(value);
  return Math.round(value).toLocaleString();
}

function formatPercent(value: number | null | undefined): string {
  return value === null || value === undefined ? '—' : `${Math.round(value)}%`;
}

function formatGigabytes(value: number): string {
  return `${(value / 1024 / 1024 / 1024).toFixed(1)}GB`;
}

function formatChartTime(value: number): string {
  return new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit' }).format(new Date(value));
}
