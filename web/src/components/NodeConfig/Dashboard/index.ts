/**
 * Dashboard Widget Editors
 *
 * Centralized export for all dashboard widget configuration editors
 */

export { ChartEditor } from './ChartEditor'
export { GaugeEditor } from './GaugeEditor'
export { FormBuilderEditor } from './FormBuilderEditor'
export { TableEditor } from './TableEditor'
export {
  TextEditor,
  ButtonEditor,
  SliderEditor,
  SwitchEditor,
  TextInputEditor,
  DropdownEditor,
  NotificationEditor,
  TemplateEditor,
} from './SimpleWidgetEditors'

export type { ChartSeries, ChartType } from './ChartEditor'
export type { GaugeSector, GaugeType } from './GaugeEditor'
export type { FormField, FormFieldType } from './FormBuilderEditor'
export type { TableColumn, ColumnType, ColumnAlign } from './TableEditor'
