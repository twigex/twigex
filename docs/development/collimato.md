# Collimato charts

A short guide to the chart code: where things are and how a chart gets
on the screen.

## Words

- **Dataset**: the data a chart uses, for example `sales`. Older code
  and screens call it a *data model*, *model* or *cube*. It is the
  same thing.
- **Option**: the object ECharts draws a chart from.

## Where things are

The chart code is in `frontend/src/utils/collimato/charts/`. Maps are
drawn differently, so their code is separate, in
`utils/collimato/map*.js` and `components/Collimato/Charts/Map/`.

```
chartTypes.js     the list of chart types and facts about each one
chartDrawing.js   picks the right code for each chart type
axisChart.js      line, bar and time bar charts
pieChart.js       pie charts
bigNumber.js      big number
timeSlots.js      the days or months along a time axis, and their labels
chartLabels.js    labels, including "(empty)" and short numbers (1.5K)
chartSettings.js  what each chart setting does
drillDown.js      drill-down
chartPalette.js   colours
```

The rest:

```
utils/collimato/chartQuery.js                    builds the query
composables/collimato/useDashboard*.js           the dashboard
composables/collimato/useChartBuilder.js         the chart editor
composables/collimato/useEcharts.js              draws a chart
components/Collimato/Widget/ChartWidget.vue      one chart on a dashboard
components/Collimato/Charts/Configurations/      the settings panels
components/Collimato/Charts/chartTypePresentation.js   icon and label per type
```

## How a chart loads on a dashboard

1. `DashboardView.vue` loads the dashboard and its charts.
2. `useDashboardFilters.js` adds the dashboard filters to each query.
3. `useDashboardQueries.js` sends the query to the server and waits
   for the answer.
4. `transformData` (`chartDrawing.js`) turns the answer into an option.
5. `ChartWidget.vue` draws it with `useEcharts.js`.

The chart editor works the same way. `useChartBuilder.js` builds the
query and `useChartEditorQuery.js` loads it.

## Settings

All settings are in `chartSettings.js`. They run at one of two times:

- **Settings that change the data** (sort bars, missing values) run
  while the option is built.
- **Settings that change the look** (margins, legend, stack) run
  just before drawing.

To add a setting: add what it does to `chartSettings.js`, then add it
to the settings panel and its text to all four languages.

## Drill-down

All of it is in `drillDown.js`:

1. When a chart is built, it remembers the real value behind each
   label. For example, "(empty)" can mean NULL or an empty string.
2. When you right-click, it finds the values for the label you
   clicked.
3. It turns them into filters for the next query.

## Good to know

- Cube sends numbers as text, like `"10.00"`.
- `null` means "no data here". A real 0 stays 0.

## Adding a chart type

Add it in three places: `chartTypes.js`, `chartDrawing.js` and
`chartTypePresentation.js`. A test fails if one is missing.

## Tests

Tests are in a `__tests__` folder next to the code they test.
