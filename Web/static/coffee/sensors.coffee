class ValueSensorDisplay extends SensorDisplay

  initialize: ->
    cont = $('<span></span>')
    @btn = @app.makeButton(
      target: cont
      text: '...'
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
