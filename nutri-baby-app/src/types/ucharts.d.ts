declare module '@qiun/ucharts' {
  interface UChartsInstance {
    setOption(options: any): void
    show(): void
    hide(): void
    removeOption(index?: number): void
    dispose(): void
  }

  interface ChartOptions {
    [key: string]: any
  }

  interface ChartData {
    [key: string]: any
  }

  function uCharts(options: {
    $canvas?: any
    canvasId?: string
    type?: string
    chartData: ChartData
    opts?: ChartOptions
  }): UChartsInstance

  export default uCharts
}
