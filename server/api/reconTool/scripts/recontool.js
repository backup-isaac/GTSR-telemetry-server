"use strict";
const params = ["Rmot", "m", "CDa", "Crr1", "Crr2", "Tmax", "Qmax", "Rline", "Vser", "VcMax", "VcMin"]
function addCommonConfigs(formData) {
  for (var param of params) {
    formData.append(param, document.getElementById(param).value);
  }
  formData.append("terrain", document.getElementById("terrain").checked);
}

function formEncodedCommonConfigs() {
  var formEncoded = "";
  for (var param of params) {
    formEncoded += "&" + param + "=" + document.getElementById(param).value;
  }
  formEncoded += "&terrain=" + document.getElementById("terrain").checked;
  return formEncoded;
}

function clearFile() {
  document.getElementById("upload").value = "";
}

function runCSV() {
  var fileUpload = document.getElementById("upload");
  var request = new XMLHttpRequest();
  request.onreadystatechange = function() {
    if (request.readyState == 4) {
      if (request.status == 200) {
        displayResult(request.response);
      } else {
        hideLoadingSpinner();
        alert("Request failed:\n" + request.responseText);
      }
    }
  }
  if (fileUpload.files.length < 1) {
    alert("Please upload at least one file.");
    return;
  }
  request.open("POST", "/reconTool/fromCSV", true);
  var formData = new FormData();
  addCommonConfigs(formData);
  formData.append("autoPlots", document.getElementById("autoPlots").checked);
  formData.append("compileFiles", document.getElementById("compileFiles").checked);
  for (var i = 0; i < fileUpload.files.length; i++) {
    formData.append("file" + i, fileUpload.files[i]);
  }
  request.send(formData);
  showLoadingSpinner();
}

function runServer() {
  var startTime = Date.parse(document.getElementById("start").value);
  var endTime = Date.parse(document.getElementById("end").value);
  var resolution = parseInt(document.getElementById("resolution").value);
  if (isNaN(startTime) || isNaN(endTime) || startTime >= endTime || (endTime - startTime) / 3600000 > 24) {
    alert("Please enter a valid start and end date with a span of no more than 24 hours.");
    return;
  }
  if (resolution < 100) {
    alert("Please enter a resolution time of at least 100ms");
    return;
  }
  var offsets = document.getElementById("timezone-offset").value.split(":").map((item) => {return parseInt(item);});
  var startDate = new Date(0);
  var endDate = new Date(0);
  /*
  How the math works:
  receivedTime = enteredTime - datetimeOffset
  trueUTC = enteredTime - selectedOffset
          = receivedTime + datetimeOffset - selectedOffset
  datetimeOffset is returned as a positive number instead of -300 which is what is expected, so it's actually subtracted
  This may or may not break across daylight savings time boundaries
  */
  startTime -= startDate.getTimezoneOffset() * 60000 + (offsets[0] * 3600000 + offsets[1] * 60000);
  endTime -= endDate.getTimezoneOffset() * 60000 + (offsets[0] * 3600000 + offsets[1] * 60000);
  var request = new XMLHttpRequest();
  request.onreadystatechange = function() {
    if (request.readyState == 4) {
      if (request.status == 200) {
        displayResult(request.response);
      } else {
        hideLoadingSpinner();
        alert("Request failed:\n" + request.responseText);
      }
    }
  }
  request.open("POST", "/reconTool/timeRange", true);
  request.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
  var formEncoded = "startDate=" + startTime + "&endDate=" + endTime + "&resolution=" + resolution;
  formEncoded += formEncodedCommonConfigs();
  request.send(formEncoded);
  showLoadingSpinner();
}

function hideLoadingSpinner() {
  document.getElementById("buttonContainer").style.display = "";
  document.getElementById("loading").style.display = "none";
}

function showLoadingSpinner() {
  document.getElementById("buttonContainer").style.display = "none";
  document.getElementById("loading").style.display = "";
}

function calculatePlotDimensions() {
  var width = window.innerWidth * 0.9;
  var height = width * 4.5 / 7;
  if (height > window.innerHeight * 0.9 && height > 300) {
    height = window.innerHeight * 0.9;
    width = height * 7 / 4.5;
  }
  return {
    width, height
  }
}

function timeSeriesTrace(data, metric, legendEntry, yAxis) {
  var trace = {
    x: data.time_min,
    y: data[metric],
    mode: "lines",
    type: "scatter",
  };
  if (legendEntry) {
    trace.name = legendEntry;
  }
  if (yAxis) {
    trace.yaxis = yAxis;
  }
  return trace;
}

function layOutAxes(chartTitle, axisTitles, showLegend, xAxis, width, height) {
  if (typeof showLegend != 'boolean') {
    showLegend = false;
  }
  if (typeof xAxis != 'string') {
    xAxis = "Time (min)"
  }
  var { width, height } = calculatePlotDimensions();
  var layout = {
    showlegend: showLegend,
    title: {
      text: chartTitle,
    },
    grid: {
      rows: axisTitles.length,
      columns: 1,
      pattern: 'independent',
    },
    xaxis: {
      title: {
        text: xAxis,
      },
    },
    width: width,
    height: height,
  };
  var domainSplit = axisTitles.length * 5.0 - 1.0;
  for (var i = 0; i < axisTitles.length; i++) {
    var dStart = i * 5.0 / domainSplit;
    var dEnd = dStart + (4.0 / domainSplit);
    layout['yaxis' + (i + 1)] = {
      domain: [dStart, dEnd],
      title: {
        text: axisTitles[i],
        font: {
          size: 13 - axisTitles.length,
        },
      },
    };
  }
  return layout;
}

function statesContainer(data) {
  var plotContainer = document.createElement("DIV");
  var kinematicsTraces = [
    timeSeriesTrace(data, 'distance_mi'),
    timeSeriesTrace(data, 'velocity_mph', null, 'y2'),
    timeSeriesTrace(data, 'acceleration', null, 'y3'),
  ];
  Plotly.newPlot(plotContainer, kinematicsTraces, layOutAxes("Vehicle States", ["Distance (mi)", "Velocity (mph)", "Acceleration (m/s^2)"]), {responsive: true});
  return plotContainer;
}

function powerTorqueContainer(data) {
  var plotContainer = document.createElement("DIV");
  var tMaxTrace = {
    x: [data.time_min[0], data.time_min[data.time_min.length - 1]],
    y: [data.max_torque, data.max_torque],
    mode: "lines",
    type: "scatter",
    line: {
      dash: "dash"
    },
    yaxis: 'y2',
    name: "Max Torque",
  };
  var powerTorqueTraces = [
    timeSeriesTrace(data, 'model_derived_torque', "Model-Derived Torque", 'y2'),
    tMaxTrace,
    timeSeriesTrace(data, 'motor_power', "Torque-Derived Power"),
    timeSeriesTrace(data, 'model_derived_power', "Model-Derived Power"),
    timeSeriesTrace(data, 'bus_power', "Measured Power"),
  ];
  Plotly.newPlot(plotContainer, powerTorqueTraces, layOutAxes("Power and Torque", ["Power (W)", "Torque (N-m)"], true));
  return plotContainer;
}

function chargeContainer(data) {
  var plotContainer = document.createElement("DIV");
  var qMaxTrace = {
    x: [data.time_min[0], data.time_min[data.time_min.length - 1]],
    y: [data.pack_capacity, data.pack_capacity],
    mode: "lines",
    type: "scatter",
    line: {
      dash: "dash"
    },
    name: "Pack Maximum",
  };
  var chargeTraces = [
    timeSeriesTrace(data, 'simulated_total_charge', "Simulated Total"),
    timeSeriesTrace(data, 'simulated_net_charge', "Simulated Net"),
    timeSeriesTrace(data, 'measured_total_charge', "Measured Out"),
    timeSeriesTrace(data, 'measured_net_charge', "Measured Net"),
    qMaxTrace,
  ];
  Plotly.newPlot(plotContainer, chargeTraces, layOutAxes("Charge Consumed", ["Charge (A-hr)"], true));
  return plotContainer;
}

function packResistanceContainer(data) {
  var plotContainer = document.createElement("DIV");
  var imin = 0;
  var imax = 0;
  for (var i of data.bms_current) {
    if (i < imin) {
      imin = i;
    } else if (i > imax) {
      imax = i;
    }
  }
  var vmax = data.pack_y_intercept - imin * data.pack_resistance;
  var vmin = data.pack_y_intercept - imax * data.pack_resistance;
  var packResistanceTraces = [
    {
      x: data.bms_current,
      y: data.bus_voltage,
      mode: "markers",
      type: "scattergl",
      // webgl rendering looks worse but SVG scatterplots murder the
      // performance of everything when you have more than a small handful
      // of points
    }, {
      x: [imin, 0, imax],
      y: [vmax, data.pack_y_intercept, vmin],
      mode: "lines",
      line: {
        width: 5,
      },
      type: "scattergl",
    },
  ];
  var title = "Pack Resistance: "
  var resStrs = (1000 * data.pack_resistance).toString().split('.')
  title += resStrs[0]
  if (resStrs.length > 1) {
    title += "." + resStrs[1].substring(0,3)
  }
  title += " mΩ"
  Plotly.newPlot(plotContainer, packResistanceTraces, layOutAxes(title, ["Bus Voltage (V)"], false, "Bus Current (A)"));
  return plotContainer
}

function batteryVoltagesContainer(data) {
  var plotContainer = document.createElement("DIV");
  var voltageTraces = [];
  var firstRaw = true;
  var maxTrace = {};
  var minTrace = {};
  for (var i = 0; i < data.module_voltages.length; i++) {
    var trace = {
      x: data.time_min,
      y: data.module_voltages[i],
      mode: "lines",
      type: "scatter",
      line: {},
      yaxis: 'y3',
    };
    if (i + 1 == data.max_module_mode) {
      trace.line.color = 'rgb(0, 170, 0)';
      trace.name = "Max";
      maxTrace = trace;
    } else if (i + 1== data.min_module_mode) {
      trace.line.color = 'rgb(255, 0, 0)';
      trace.name = "Min";
      minTrace = trace;
    } else {
      trace.line.color = 'rgb(204, 204, 204)';
      if (firstRaw) {
        trace.name = "Raw";
        firstRaw = false;
      } else {
        trace.showlegend = false;
      }
      voltageTraces.push(trace);
    }
  }
  voltageTraces.push(
    maxTrace,
    minTrace,
    timeSeriesTrace(data, 'max_min_difference', "Max-Min Difference", 'y2'),
    timeSeriesTrace(data, 'max_module', "Max Module"),
    timeSeriesTrace(data, 'min_module', "Min Module")
  );
  Plotly.newPlot(plotContainer, voltageTraces, layOutAxes("Module Voltages", ["Min/Max Module", "Max-Min (V)", "Module Voltage (V)"], true));
  return plotContainer;
}

function packStatsContainer(data) {
  var plotContainer = document.createElement("DIV");
  var milliohms = data.module_resistances.map((item) => { return item * 1000; });
  var rMin = Math.min.apply(null, milliohms);
  var rMax = Math.max.apply(null, milliohms);
  var mu = data.mean_module_resistance * 1000;
  var sugma = data.module_standard_deviation * 1000;
  var pdfX = [];
  var pdfY = [];
  var packTraces = [{
    type: 'bar',
    x: data.module_resistances.map((_, i) => { return i + 1; }),
    y: milliohms,
    xaxis: 'x2',
    yaxis: 'y2',
    showlegend: false,
  }, {
    type: 'histogram',
    x: milliohms,
    name: "Frequency",
  }];
  var sum = 0;
  for (var res of data.module_resistances) {
    sum += res * 1000;
  }
  var titleText = "Module Resistance and Distribution - ";
  titleText += sum.toString().split('.')[0]
  titleText += " mΩ Total"
  var { width, height } = calculatePlotDimensions();
  var layout = {
    showlegend: true,
    title: {
      text: titleText,
    },
    grid: {
      rows: 2,
      columns: 1,
      pattern: 'independent',
    },
    xaxis1: {
      title: {
        text: "Module Resistance (mΩ)",
      },
    },
    xaxis2: {
      anchor: 'y2',
      title: {
        text: "Module Number",
        standoff: 0,
      },
    },
    yaxis1: {
      domain: [0, 4.0/9],
      title: {
        text: "# of Modules",
        font: {
          size: 11,
        },
      },
    },
    yaxis2: {
      domain: [5.0/9, 1],
      title: {
        text: "Module Resistance (mΩ)",
        font: {
          size: 11,
        },
      },
    },
    annotations: [{
      x: mu - 3 * sugma,
      y: 0,
      text: "-3σ",
      showarrow: false,
      yshift: -6
    }, {
      x: mu + 3 * sugma,
      y: 0,
      text: "+3σ",
      showarrow: false,
      yshift: -6
    }],
    width: width,
    height: height,
  };
  Plotly.newPlot(plotContainer, packTraces, layout /*layOutAxes(null, ["# of Modules", "Module Resistance (mΩ)"], false, "Module Number")*/);
  // going deep into the internals of Plotly histogram
  // this is actually how the Plotly engineers recommend getting the
  // bins/bin characteristics of a histogram that Plotly makes
  // no functions are exposed for this purpose
  var histMax = plotContainer._fullData[1]._extremes.y.max[0].val;
  for (var i = Math.min(rMin, mu - 3 * sugma); i < Math.max(rMax, mu + 3 * sugma); i += (rMax - rMin) / 420) {
    pdfX.push(i);
    var pdfi = histMax * Math.exp(-1 * ((i - mu) / sugma) * ((i - mu) / sugma) / 2);
    pdfY.push(pdfi);
  }
  packTraces.push({
    x: pdfX,
    y: pdfY,
    mode: "lines",
    type: "scatter",
    name: "Normal Distribution"
  });
  layout.shapes = [{
    type: 'line',
    x0: mu - 3 * sugma,
    x1: mu - 3 * sugma,
    y0: 0,
    y1: histMax * 1.1,
    line: {
      dash: "dash",
      color: "rgb(255, 0, 0)",
    },
  }, {
    type: 'line',
    x0: mu + 3 * sugma,
    x1: mu + 3 * sugma,
    y0: 0,
    y1: histMax * 1.1,
    line: {
      dash: "dash",
      color: "rgb(255, 0, 0)",
    },
  }];
  Plotly.react(plotContainer, packTraces, layout);
  return plotContainer;
}

function drivetrainEfficiencyContainer(data) {
  var plotContainer = document.createElement("DIV");
  var effTraces = [
    timeSeriesTrace(data, 'drivetrain_efficiency', "Total Drivetrain Efficiency"),
    timeSeriesTrace(data, 'motor_efficiency', "Motor Efficiency"),
    timeSeriesTrace(data, 'mc_efficiency', "Motor Controller Efficiency"),
    timeSeriesTrace(data, 'pack_efficiency', "Battery Pack Efficiency"),
  ];
  Plotly.newPlot(plotContainer, effTraces, layOutAxes("Drivetrain Efficiency and Components", ["Efficiency (%)"], true));
  return plotContainer;
}

function solarContainer(data) {
  var plotContainer = document.createElement("DIV");
  var solarTraces = [
    timeSeriesTrace(data, 'solar_power', null, 'y2'),
    timeSeriesTrace(data, 'solar_charge'),
  ];
  Plotly.newPlot(plotContainer, solarTraces, layOutAxes("Solar Array Power and Charge", ["Charge (A-hr)", "Power (W)"], false));
  return plotContainer;
}

function vtpContainer(data) {
  var plotContainer = document.createElement("DIV");
  var vtpTraces = [
    timeSeriesTrace(data, 'velocity_mph', null, 'y3'),
    timeSeriesTrace(data, 'motor_torque', null, 'y2'),
    timeSeriesTrace(data, 'bus_power'),
  ];
  Plotly.newPlot(plotContainer, vtpTraces, layOutAxes("VTP Trajectories", ["Power (W)", "Torque (N-m)", "Velocity (mph)"], false));
  return plotContainer;
}

function tptContainer(data) {
  // jackson calls this "torque, phase, throttle..."
  // but is actually "velocity, phase, throttle"
  var plotContainer = document.createElement("DIV");
  var vtpTraces = [
    timeSeriesTrace(data, 'velocity_mph', null, 'y3'),
    timeSeriesTrace(data, 'phase_current', null, 'y2'),
    timeSeriesTrace(data, 'throttle'),
  ];
  Plotly.newPlot(plotContainer, vtpTraces, layOutAxes("Velocity, Phase, and Throttle", ["Throttle (%)", "Phase Current (Arms)", "Velocity (mph)"], false));
  return plotContainer;
}

function speedContourContainer(data) {
  var plotContainer = document.createElement("DIV");
  var speedContourTrace = [{
    x: data.x_disp,
    y: data.y_disp,
    mode: "markers",
    type: "scattergl",
    marker: {
      color: data.velocity_mph,
    },
  }];
  Plotly.newPlot(plotContainer, speedContourTrace, layOutAxes("Race Route Speed Contour", ["Distance (mi)"], false, "Distance (mi)"));
  return plotContainer;
}

function rawPlotContainer(data, rawValue) {
  var plotContainer = document.createElement("DIV");
  Plotly.newPlot(plotContainer, [{
    x: data.raw_timestamps,
    y: data.raw_values[rawValue],
    mode: "lines",
    type: "scatter",
  }], layOutAxes(`${rawValue} versus time`, [rawValue], false));
  return plotContainer;
}

function toggleAutoPlots(index) {
  var rawPlots = document.getElementById(`rawPlotsContainer${index}`);
  var toggleButton = document.getElementById(`toggleAutoPlotsButton${index}`);
  if (rawPlots.style.display != "none") {
    toggleButton.innerHTML = "Show Raw Plots";
    rawPlots.style.display = "none";
  } else {
    toggleButton.innerHTML = "Hide Raw Plots";
    rawPlots.style.display = "block";
  }
}

var autoPlotsCounter = 0;

function autoPlotsContainer(data) {
  if (data.raw_values && data.raw_timestamps) {
    var autoPlotsContainer = document.createElement("DIV");
    autoPlotsContainer.className = "panel panel-default";
    var heading = document.createElement("DIV");
    heading.className = "panel-heading";
    var body = document.createElement("DIV");
    body.class = "panel-body";
    body.id = `rawPlotsContainer${autoPlotsCounter}`;
    var buttonHolder = document.createElement("DIV");
    buttonHolder.className = "panel-body";
    buttonHolder.style.textAlign = "center";
    buttonHolder.innerHTML = `<button id="toggleAutoPlotsButton${autoPlotsCounter}" onclick="toggleAutoPlots(${autoPlotsCounter})" class="btn btn-primary">Hide Raw Plots</button>`;
    var promise = new Promise((resolve, reject) => {
      var container = body;
      var start = data.raw_timestamps[0];
      var rawMinutes = [];
      data.raw_timestamps.forEach((rawTime) => {
        rawMinutes.push((rawTime - start) / 60000);
      });
      data.raw_timestamps = rawMinutes;
      resolve({container, data});
    });
    for (var rawValue in data.raw_values) {
      promise = promise.then(rawPromiseCallback(rawValue));
    }
    autoPlotsContainer.appendChild(heading);
    heading.innerHTML = '<h2 style="text-align: center">Raw Plots</h2>';
    autoPlotsContainer.appendChild(buttonHolder);
    autoPlotsContainer.appendChild(body);
    autoPlotsCounter++;
    return autoPlotsContainer;
  }
  return null;
}


function rawPromiseCallback(rawValue) {
  return function(obj) {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        var { container, data } = obj;
        container.appendChild(rawPlotContainer(data, rawValue));
        resolve(obj);
      }, 0);
    });
  }
}

function displayResult(response) {
  document.getElementById("inputs").style.display = "none";
  document.getElementById("outputs").style.display = "inline";
  var bigData = JSON.parse(response)
  var { width, height } = calculatePlotDimensions();
  var oldWidth = width;
  if (width < 420) {
    document.getElementById('smallWarning').style.display = "block";
    document.getElementById('results').style.display = "none";
  }
  window.onresize = function() {
    if (document.getElementById('outputs').style.display == "none") {
      return;
    }
    var { width, height } = calculatePlotDimensions();
    if (width < 420 && oldWidth >= 420) {
      document.getElementById('smallWarning').style.display = "block";
      document.getElementById('results').style.display = "none";
    } else if (width < 420) {
      return;
    }
    if (width >= 420 && oldWidth < 420) {
      document.getElementById('smallWarning').style.display = "none";
      document.getElementById('results').style.display = "block";
    }
    if (oldWidth == width) {
      return;
    }
    for (var plotContainer of document.getElementById('results').childNodes) {
      if (plotContainer.id && plotContainer.id.startsWith('plot')) {
        for (var div of plotContainer.childNodes) {
          if (div.className == "js-plotly-plot") {
            Plotly.relayout(div, {
              width: width,
              height: height,
            });
          }
        }
      }
    }
    oldWidth = width;
  }
  for (var i = 0; i < bigData.length; i++) {
    var plotContainer = document.createElement("DIV");
    plotContainer.style.display = "block";
    plotContainer.id = "plot" + i;
    document.getElementById("results").appendChild(plotContainer);
    if (i != bigData.length - 1) {
      var div = document.createElement("DIV");
      div.style.minHeight = "5px";
      div.style.display = "block";
      div.style.backgroundColor = "rgb(224, 224, 224)";
      document.getElementById("results").appendChild(div);
    }
    new Promise((resolve, reject) => {
      plotContainer.appendChild(statesContainer(bigData[i]));
      var container = plotContainer;
      var data = bigData[i];
      resolve({container, data});
    }).then(promiseCallback(powerTorqueContainer))
      .then(promiseCallback(chargeContainer))
      .then(promiseCallback(packResistanceContainer))
      .then(promiseCallback(batteryVoltagesContainer))
      .then(promiseCallback(packStatsContainer))
      .then(promiseCallback(drivetrainEfficiencyContainer))
      .then(promiseCallback(solarContainer))
      .then(promiseCallback(vtpContainer))
      .then(promiseCallback(tptContainer))
      .then(promiseCallback(speedContourContainer))
      .then(promiseCallback(autoPlotsContainer));
  }
  document.getElementById("downloadRawButton").onclick = function() {
    download(response, "recontool_raw.json", "application/json")
  }
}

function promiseCallback(containerFunction) {
  return function(obj) {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        var { container, data } = obj;
        var childContainer = containerFunction(data);
        if (childContainer) {
          container.appendChild(childContainer);
        }
        resolve(obj);
      }, 0);
    });
  }
}

function goBack() {
  hideLoadingSpinner();
  document.getElementById("inputs").style.display = "inline";
  document.getElementById("results").innerHTML = "";
  document.getElementById("outputs").style.display = "none";
}

function download(fileContents, filename, mimeType) {
  var file = new Blob([fileContents], {type: mimeType});
  saveAs(file, filename);
}

function setupPickers() {
  new Picker(document.getElementById('start'), {
    format: 'YYYY/MM/DD HH:mm',
    controls: true,
  });
  new Picker(document.getElementById('end'), {
    format: 'YYYY/MM/DD HH:mm',
    controls: true,
  });
}