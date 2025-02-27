import React, { useEffect, useRef, useState } from "react";
import {
    createChart,
    ColorType,
    IChartApi,
    ISeriesApi,
    CandlestickSeries,
    LineSeries,
} from "lightweight-charts";

export const ChartComponent: React.FC = () => {
    const chartContainerRef = useRef<HTMLDivElement | null>(null);
    const chartRef = useRef<IChartApi | null>(null);
    const rsiChartRef = useRef<IChartApi | null>(null);
    const [data, setData] = useState<any[]>([]);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const response = await fetch("/rsi_bb.json"); // Ensure correct path
                const jsonData = await response.json();

                // Ensure data is sorted and properly formatted
                const formattedData = jsonData
                    .map((d: any) => ({
                        time: d.time, // Ensure time is in seconds format (Unix timestamp)
                        open: parseFloat(d.open),
                        high: parseFloat(d.high),
                        low: parseFloat(d.low),
                        close: parseFloat(d.close),
                        upperBB: d.Upper_BB ? parseFloat(d.Upper_BB) : null,
                        lowerBB: d.Lower_BB ? parseFloat(d.Lower_BB) : null,
                        rsi: d.RSI ? parseFloat(d.RSI) : null,
                        longEntry: d.Long_Entry,
                        shortEntry: d.Short_Entry,
                        longExit: d.Long_Exit,
                        shortExit: d.Short_Exit,
                    }))
                    .filter((d) => !isNaN(d.time) && !isNaN(d.open)) // Remove invalid entries
                    .sort((a, b) => a.time - b.time); // Ensure sorted order

                setData(formattedData);
            } catch (error) {
                console.error("Error loading JSON:", error);
            }
        };

        fetchData();
    }, []);

    useEffect(() => {
        if (!chartContainerRef.current || data.length === 0) return;

        const chart = createChart(chartContainerRef.current, {
            layout: {
                background: { type: ColorType.Solid, color: "white" },
                textColor: "black",
            },
            width: chartContainerRef.current.clientWidth * 0.9,
            height: (chartContainerRef.current.clientHeight * 0.6),
        });

        chart.timeScale().fitContent();
        chartRef.current = chart;

        const candlestickSeries = chart.addSeries(CandlestickSeries, {
            upColor: "#26a69a",
            downColor: "#ef5350",
            borderVisible: false,
            wickUpColor: "#26a69a",
            wickDownColor: "#ef5350",
        });

        candlestickSeries.setData(data);

        // Bollinger Bands Lines
        const upperBBSeries = chart.addSeries(LineSeries, {
            color: "#ff9900",
            lineWidth: 2,
        });
        upperBBSeries.setData(
            data.filter((d) => d.upperBB !== null).map(({ time, upperBB }) => ({ time, value: upperBB }))
        );

        const lowerBBSeries = chart.addSeries(LineSeries, {
            color: "#ff9900",
            lineWidth: 2,
        });
        lowerBBSeries.setData(
            data.filter((d) => d.lowerBB !== null).map(({ time, lowerBB }) => ({ time, value: lowerBB }))
        );

        // Adding RSI Subplot
        const rsiChart = createChart(chartContainerRef.current, {
            layout: {
                background: { type: ColorType.Solid, color: "white" },
                textColor: "black",
            },
            width: chartContainerRef.current.clientWidth * 0.9,
            height: (chartContainerRef.current.clientHeight * 0.3),
        });
        rsiChart.timeScale().fitContent();
        rsiChartRef.current = rsiChart;

        const rsiSeries = rsiChart.addSeries(LineSeries, {
            color: "#0000FF",
            lineWidth: 2,
        });
        rsiSeries.setData(
            data.filter((d) => d.rsi !== null).map(({ time, rsi }) => ({ time, value: rsi }))
        );

        const handleResize = () => {
            if (chartContainerRef.current) {
                chart.applyOptions({
                    width: chartContainerRef.current.clientWidth * 0.9,
                    height: chartContainerRef.current.clientHeight * 0.6,
                });
                rsiChart.applyOptions({
                    width: chartContainerRef.current.clientWidth * 0.9,
                    height: chartContainerRef.current.clientHeight * 0.3,
                });
            }
        };

        window.addEventListener("resize", handleResize);

        return () => {
            window.removeEventListener("resize", handleResize);
            chart.remove();
            rsiChart.remove();
        };
    }, [data]);

    return (
        <div
            style={{
                display: "flex",
                flexDirection: "column",
                justifyContent: "center",
                alignItems: "center",
                width: "100vw",
                height: "100vh",
                backgroundColor: "#f4f4f4",
            }}
        >
            <div ref={chartContainerRef} style={{ width: "90%", height: "90%" }} />
        </div>
    );
};
