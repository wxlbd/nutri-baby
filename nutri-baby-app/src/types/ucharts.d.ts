declare module '@qiun/ucharts' {
  interface UChartsConstructorOptions {
    type?: string
    context?: CanvasRenderingContext2D
    canvasId?: string
    width?: number
    height?: number
    categories?: string[]
    series?: Array<{
      name: string
      data: any[]
    }>
    animation?: boolean
    pixelRatio?: number
    color?: string[]
    padding?: [number, number, number, number]
    legend?: any
    xAxis?: any
    yAxis?: any
    extra?: any
    opts?: any
    [key: string]: any
  }

  interface UChartsInstance {
    setOption(options: any): void
    updateData(options: any): void
    show(): void
    hide(): void
    removeOption(index?: number): void
    dispose(): void
    touchLegend?(touch: any): void
    [key: string]: any
  }

  interface ChartOptions {
    [key: string]: any
  }

  interface ChartData {
    [key: string]: any
  }

  class uCharts {
    constructor(options: UChartsConstructorOptions)
    setOption(options: any): void
    updateData(options: any): void
    show(): void
    hide(): void
    removeOption(index?: number): void
    dispose(): void
    touchLegend?(touch: any): void
    [key: string]: any
  }

  export default uCharts
}

