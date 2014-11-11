normalize = (a) ->
  n = a.length-1
  if n < 7 then return a
  result = []
  for val, i in a
    v = val
    switch i
      when 0
        v = (39*a[0] + 8*a[1] - 4*(a[2]+a[3]-a[4]) + a[5] - 2*a[6]) / 42
      when 1
        v = (8*a[0] + 19*a[1] + 16*a[2] + 6*a[3] - 4*a[4] - 7*a[5] + 4*a[6]) / 42
      when 2
        v = (-4*a[0] + 16*a[1] + 19*a[2] + 12*a[3] + 2*a[4] - 4*a[5] + 1*a[6]) / 42
      when n-2
        v = (-4*a[n] + 16*a[n-1] + 19*a[n-2] + 12*a[n-3] + 2*a[n-4] - 4*a[n-5] + 1*a[n-6]) / 42
      when n-1
        v = (8*a[n] + 19*a[n-1] + 16*a[n-2] + 6*a[n-3] - 4*a[n-4] - 7*a[n-5] + 4*a[n-6]) / 42
      when n
        v = (39*a[n] + 8*a[n-1] - 4*a[n-2] - 4*a[n-3] + a[n-4] + 4*a[n-5] - 2*a[n-6]) / 42
      else
        v = (7*a[i] + 6*a[i+1] + 6*a[i-1] + 3*a[i+2] + 3*a[i-2] - 2*a[i+3] - 2*a[i-3]) / 21
    result.push(v)
  return result

class ValueSensorDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @btn = @app.makeButton(
      target: cont
      text: '...'
      handler: =>
        @room.showDetail(@extra.detail) if @extra.detail
    )
    @app.addDataListener(@config, (data) =>
      @show(data.value)
      # Refresh
    )
    return cont

  show: (value) ->
    text = ''
    if @extra.percent
      text = "#{Math.round(value)}%"
    if @extra.temp
      text = "#{Math.round(value*10) / 10}°"
    @btn.text(text)

registerSensor('sensor_value', ValueSensorDisplay)

class DateSensorDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @btn = @app.makeButton(
      target: cont
      text: '...'
    )
    return cont

  refresh: ->
    format = ''
    if @extra.time
      format = 'HH:mm'
    if @extra.date
      format = 'dddd, MMMM d'
    dt = new Date()
    @btn.text(dt.format(format))

registerSensor('date', DateSensorDisplay)

class DateTimeSensorDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @contents = $("""
    <div class="sensor_dt_root">
      <div class="sensor_dt_time text-large"></div>
      <div class="sensor_dt_date">
        <div class="sensor_dt_day text-large"></div>
        <div class="sensor_dt_week_month">
          <div class="sensor_dt_week text-small"></div>
          <div class="sensor_dt_month text-small"></div>
        </div>
      </div>
    </div>""")
    @btn = @app.makeButton(
      target: cont
      cls: 'no-height'
      contents: @contents
    )
    return cont

  refresh: ->
    format = ''
    dt = new Date()
    @contents.find('.sensor_dt_time').text(dt.format('HH:mm'))
    @contents.find('.sensor_dt_day').text(dt.format('d'))
    @contents.find('.sensor_dt_week').text(dt.format('dddd'))
    @contents.find('.sensor_dt_month').text(dt.format('MMMM'))

registerSensor('date_time', DateTimeSensorDisplay)

class WeatherIconDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @btn = @app.makeButton(
      target: cont
      handler: =>
        @room.showDetail(@extra.detail) if @extra.detail
    )
    @app.addDataListener(@config, (data) =>
      @show(data.value)
    )
    return cont

  show: (value) ->
    icon = 'sun'
    switch value
      when 2 then icon = 'pcloudy' # patly cloudy
      when 3 then icon = 'cloudy' # cloudy
      when 4 then icon = 'rain' # rain
      when 5 then icon = 'lightning' # tstorms
      when 6 then icon = 'snow' # show
      when 7 then icon = 'fog' # fog
    @btn.html("<span class='fa icon-weather icon-#{icon}'></span>")

registerSensor('weather_icon', WeatherIconDisplay)

class LightDisplay extends SensorDisplay

  initialize: (@roomDiv) ->
    @app.addDataListener(@config, (data) =>
      @show(data.value)
    )
    return undefined

  show: (value) ->
    min = @extra.min
    max = @extra.max
    if value < min
      value = min
    if value > max
      value = max
    intensivity = Math.ceil((max - value) / (max - min) * 255)
    # log 'Light', value, min, max, intensivity
    @roomDiv.css(
      'background-color': "rgb(#{intensivity}, #{intensivity}, #{intensivity})"
    )

registerSensor('light', LightDisplay)

class SensorFlagDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @btn = @app.makeButton(
      target: cont
      icon: @extra.icon
    )
    @app.addDataListener(@config, (data) =>
      @show(data.value)
    )
    @show(0)
    return cont

  show: (value) ->
    # log 'show', @extra
    @btn.almostHide(value == 1)

registerSensor('sensor_flag', SensorFlagDisplay)

class SensorValueDirectionDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @contents = $('<span class="no-wrap"><i class="fa fa-arrow-down"></i><span class="text">...</span></span>')
    @btn = @app.makeButton(
      target: cont
      contents: @contents
    )
    @app.addDataListener(@config, (data) =>
      @value = data.value
    )
    @app.addDataListener(
      device: @config.device
      type: @config.type
      index: @config.index
      measure: @extra.dir
    , (data) =>
      @show(data.value, @value)
    )
    return cont

  show: (dir, value) ->
    # log 'Wind:', dir, value
    @contents.find('.text').text("#{value}")
    @contents.find('i').css(
      'transform': "rotate(#{dir}deg)"
      '-webkit-transform': "rotate(#{dir}deg)"
    )

registerSensor('direction_value', SensorValueDirectionDisplay)

prepareSeries = (data, doNormalize) ->
    arr = []
    min = 9999
    max = 0
    raw = []
    for item in data
      arr.push([item.ts, item.value])
      raw.push(item.value)
    norm = raw
    normTimes = Math.ceil(raw.length / 50)
    if doNormalize and normTimes > 0
      for j in [0...normTimes]
        norm = normalize(norm)
    for val, j in norm
      if val < min then min = val
      if val > max then max = val
      arr[j][1] = val
    return [arr, min, max]

class InlineGraphDisplay extends SensorDisplay

  COLORS: ['#dc322f', '#268bd2', '#859900']

  initialize: ->

  refresh: ->
    to = new Date().getTime()
    from = to - 1 * 24 * 60 * 60 * 1000 # 3 days
    @app.fetchData(@config.data, from, to).then((data) =>
      series = []
      yaxes = []
      colors = []
      for data, i in data.series
        [arr, min, max] = prepareSeries(data, yes)
        series.push(
          data: arr
          yaxis: i+1
        )
        gap = (max - min) / 2
        yaxes.push({
          min: min-gap
          max: max+gap
        })
        colors.push(@COLORS[i % @COLORS.length])
      @room.plot(series, colors, yaxes)
    )

registerSensor('inline_graph', InlineGraphDisplay)

COLORS =
  'yellow':  [0xb5, 0x89, 0x00]
  'orange':  [0xcb, 0x4b, 0x16]
  'red':     [0xdc, 0x32, 0x2f]
  'magenta': [0xd3, 0x36, 0x82]
  'violet':  [0x6c, 0x71, 0xc4]
  'blue':    [0x26, 0x8b, 0xd2]
  'cyan':    [0x2a, 0xa1, 0x98]
  'green':   [0x85, 0x99, 0x00]
  'grey':    [0x83, 0x94, 0x96]

class DetailGraphDisplay extends SensorDisplay
 
  initialize: ->
    renderOne = (idx) =>
      from = control.from.getTime()
      to = control.to.getTime()
      @app.fetchData(@config.data[idx].sensors, from, to, @extra.forecast).then((data) =>
        if not control.visible then return
        series = []
        yaxes = []
        colors = []
        yaxes.push({
          min: 0
          max: 100
          show: yes
          position: "right"
          labelWidth: 30
        })
        for data, i in data.series
          conf = @config.data[idx].sensors[i]
          [arr, min, max] = prepareSeries(data, yes)
          series.push(
            data: arr
            lines:
              show: yes
              fill: conf.fill
            yaxis: if conf.percent then 1 else yaxes.length+1
          )
          gap = (max - min) / 2
          if not conf.percent
            yaxes.push({
              labelWidth: 50
              min: if min-gap > 0 then min-gap else 0
              max: max+gap
            })
          col = COLORS[conf.color] ? COLORS.yellow
          colors.push("rgb(#{col[0]}, #{col[1]}, #{col[2]})")
        control.plot(control.divs[idx], series, colors, yaxes)
      )
    control = new DetailsDialog(
      app: @app
      forecast: @extra.forecast
      default: @extra.default
      onCreate: (div) =>
        control.divs = []
        for item in @config.data
          control.divs.push($("<div class='surface-graph'></div>").appendTo(div))
      onRender: =>
        for item, idx in @config.data
          renderOne(idx)

    )
    @room.addDetail(@extra.name, control)

  show: (value) ->

registerSensor('graph', DetailGraphDisplay)
