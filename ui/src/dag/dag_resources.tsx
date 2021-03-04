import {
  AreaChart,
  Line,
  LineChart,
  XAxis,
  YAxis,
  Area,
  Tooltip,
  CartesianGrid,
} from "recharts";
import { useParams } from "react-router";
import { DagPropName } from "./dag_page";
import { fetchDAGMetrics } from "../backend/fetch_calls";
import { useComponentWillMount } from "../hooks/component_will_mount";
import { useState } from "react";
import { DateTime } from "luxon";

type DagMetric = {
  Time: DateTime;
  CPU: number;
  Memory: number;
};

type incomingMetrics = {
  ID: number;
  DagName: string;
  PodName: string;
  Memory: number;
  CPU: number;
  MetricTime: string;
};

type AreaChartProps = {
  MetricsArray: DagMetric[];
  Attribute: string;
};

function MetricAreaChart(props: AreaChartProps) {
  return (
    <div>
      <h2>{props.Attribute}</h2>
      <AreaChart
        width={600}
        height={250}
        data={props.MetricsArray}
        margin={{ top: 10, right: 30, left: 30, bottom: 0 }}
      >
        <defs>
          <linearGradient id="colorUv" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8} />
            <stop offset="95%" stopColor="#8884d8" stopOpacity={0} />
          </linearGradient>
          <linearGradient id="colorPv" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#82ca9d" stopOpacity={0.8} />
            <stop offset="95%" stopColor="#82ca9d" stopOpacity={0} />
          </linearGradient>
        </defs>
        <XAxis
          dataKey="Time"
          type="number"
          tickFormatter={(unixTime: number) =>
            DateTime.fromMillis(unixTime).toISO()
          }
          allowDataOverflow={true}
          domain={["dataMin", "dataMax"]}
        />
        <YAxis />
        <CartesianGrid strokeDasharray="3 3" />
        <Tooltip />
        <Line type="monotone" dataKey="Memory" stroke="#8884d8" />
        <Area
          type="monotone"
          dataKey={props.Attribute}
          stroke="#8884d8"
          fillOpacity={1}
          fill="url(#colorUv)"
        />
        <Area
          type="monotone"
          dataKey="pv"
          stroke="#82ca9d"
          fillOpacity={1}
          fill="url(#colorPv)"
        />
      </AreaChart>
    </div>
  );
}

export function DAGResourceUsageBody() {
  let { name } = useParams<DagPropName>();
  const [metricsArray, setMetricsArray] = useState<DagMetric[]>([]);

  useComponentWillMount(() => {
    fetchDAGMetrics(name).then((data) => {
      let newMetricArray: DagMetric[] = [];
      data.forEach((metricDataPoint: incomingMetrics) => {
        let displayMetric = {
          Time: DateTime.fromISO(metricDataPoint.MetricTime),
          CPU: metricDataPoint.CPU,
          Memory: metricDataPoint.Memory,
        } as DagMetric;
        newMetricArray.push(displayMetric);
      });
      setMetricsArray(newMetricArray);
    });
  });

  return (
    <div>
      <MetricAreaChart MetricsArray={metricsArray} Attribute={"Memory"} />
      <MetricAreaChart MetricsArray={metricsArray} Attribute={"CPU"} />
    </div>
  );
}
