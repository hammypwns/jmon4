$(function() {

    data = [];
    trace = {
        x: [],
        y: [],
        name: 'Placeholder',
        type: 'scatter'
    };
    let dataurlPass = "/system/test/data?p";
    let dataurlFail = "/system/test/data?f";
    var dataurlSuffix = 'm';

    data.push(trace);
    Plotly.newPlot("placeholder", data);

    function plotData() {
        data = [];
        function onDataReceived(series) {
            trace = {
                x: series["data"].map((tuple) => {return tuple[0]}),
                y: series["data"].map((tuple) => {return tuple[1]}),
                name: series["label"],
                type: 'scatter'
            };
            console.log(series);
            console.log(trace);
            if (trace.name === "Passes"){
                data[0] = trace;
            }
            else{
                data[1] = trace;
            }
            Plotly.newPlot("placeholder", data);
        }

        // Get both passes and failures together
        [dataurlPass, dataurlFail].forEach((url) => {
            $.ajax({
                url: url + dataurlSuffix,
                type: "GET",
                dataType: "json",
                success: onDataReceived
            });
        })
        
    };
    
    $("button.minuteSeries").click(function () {
        getConfigData();
        dataurlSuffix = 'm'
        plotData();
    });

    $("button.hourSeries").click(function () {
        dataurlSuffix = 'h'
        plotData();
    });
    
    $("button.daySeries").click(function () {
        dataurlSuffix = 'd'
        plotData();
    });
    
    // Load the first series by default, so we don't have an empty plot
    $("button.minuteSeries").click();

});

function getConfigData(){
    var xmlhttp;
    xmlhttp=new XMLHttpRequest();
    xmlhttp.onreadystatechange=function() {
        if (xmlhttp.readyState==4 && xmlhttp.status==200) {
            document.getElementById("users-contain").innerHTML=xmlhttp.responseText;
        }
    }
    xmlhttp.open("POST","ajax/getInfo",true);
    xmlhttp.setRequestHeader("Content-type","application/x-www-form-urlencoded");
    xmlhttp.send("product=test");		
}	