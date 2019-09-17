angular.module('app', [])
    .controller('mainCtrl', ['$scope', '$http', mainCtrl]);

function mainCtrl($scope, $http) {

    // Url to both passes and failures
    $scope.dataUrl = {
        pass: "/system/test/data?p",
        fail: "/system/test/data?f"
    };

    // Suffix for minute (m), hour (h), or day (d)
    $scope.urlSuffix = 'm';

    // Gets data at current mode and plots it
    $scope.plotData = () => {
        plotdata = [];
        onDataReceived = (series) => {
            trace = {
                x: series["data"].map((tuple) => { return tuple[0] }),
                y: series["data"].map((tuple) => { return tuple[1] }),
                name: series["label"],
                type: 'scatter'
            };
            if (trace.name === "Passes") {
                plotdata[0] = trace;
            }
            else {
                plotdata[1] = trace;
            }
            Plotly.newPlot("placeholder", plotdata);
        }

        $http.get($scope.dataUrl.pass + $scope.urlSuffix).success(onDataReceived);
        $http.get($scope.dataUrl.fail + $scope.urlSuffix).success(onDataReceived);
    }

    // When minute, hour, or day button is clicked it passes in the suffix and reloads the plot
    $scope.setGraphMode = (suffix) => {
        $scope.urlSuffix = suffix;
        $scope.plotData();
    }

    // The info for the table
    $scope.tableInfo = {
        test: "N/A",
        days: "N/A",
        time: "N/A",
        passed: "N/A",
        failed: "N/A",
    }

    // Gets the data for the table at top of page
    $scope.getConfigData = () => {
        $http.get('/ajax/getInfo').success((data) => {
            angular.copy(data, $scope.tableInfo);
            window.document.title = $scope.tableInfo.test;
        })
    }

    // Stuff for setting the name of the test
    $scope.editShown = false;

    $scope.toggleEdit = () => { $scope.editShown = !$scope.editShown }

    $scope.setInfo = (name) => {
        $http.get('/setInfo?' + name).success((data) => {
            $scope.editShown = false;
            $scope.getConfigData();
        })
    }

    // Table of vm ips
    $scope.vmTable = [];

    // Get vmTable
    $scope.getVmTable = () => {
        $http.get('/vmLinks').success((res) => {
            angular.copy(res.data, $scope.vmTable);
        })
    }

    // Send vmTable
    $scope.postVmTable = () => {
        $http.post('/vmLinks', { data: $scope.vmTable }).success((res) => {
            $scope.getVmTable();
        })
    }

    // ---------------------------------------------------------------------------------------------
    // Contains notes data
    $scope.notesTable = [];

    // Get notesTable and display data
    $scope.getNotesTable = () => {
        $http.get('/notes').success((res) => {
            angular.copy(res.notes, $scope.notesTable);
            document.getElementById("notesTextArea").value = $scope.notesTable;
        })
    }

    // Send notesTable
    $scope.postNotesTable = (notesContents) => {
        $scope.notesTable[0] = notesContents;
        $http.post('/notes', { notes: $scope.notesTable }).success((res) => {
            $scope.getNotesTable();
        })
    }

    //Creates a delay after editing the textarea. Without a delay some inputs get dropped
    var typingTimer;                //timer identifier
    var doneTypingInterval = 500;  //time in ms .5 seconds
    var $input = $('#notesTextArea');

    //on keyup, start the countdown
    $input.on('keyup paste', function () {
        clearTimeout(typingTimer);
        typingTimer = setTimeout(doneTyping, doneTypingInterval);
    });

    //on keydown, clear the countdown 
    $input.on('keydown', function () {
        clearTimeout(typingTimer);
    });

    //user is "finished typing," do something
    function doneTyping() {
        var notesContents = document.getElementById("notesTextArea").value
        $scope.postNotesTable(notesContents);
    }

    //Allows for tabs in the notes textarea
    $(document).delegate('#notesTextArea', 'keydown', function(e) {
        var keyCode = e.keyCode || e.which;
      
        if (keyCode == 9) {
          e.preventDefault();
          var start = this.selectionStart;
          var end = this.selectionEnd;
      
          // set textarea value to: text before caret + tab + text after caret
          $(this).val($(this).val().substring(0, start)
                      + "\t"
                      + $(this).val().substring(end));
      
          // put caret at right position again
          this.selectionStart =
          this.selectionEnd = start + 1;
        }
      });
    // ---------------------------------------------------------------------------------------------

    $scope.addRowShown = false;
    $scope.toggleAddRow = (i) => { $scope.addRowShown = !$scope.addRowShown }

    // Add row
    $scope.addRow = (newRow) => {
        $scope.vmTable.push(newRow);
        $scope.addRowShown = false;
        $scope.postVmTable();
        document.getElementById('new-row-form').reset();
    }

    // Move VmTable row up
    $scope.moveRowUp = function (rowIndex) {
        if (rowIndex > 0) {
            tmp = $scope.vmTable[rowIndex - 1];
            $scope.vmTable[rowIndex - 1] = $scope.vmTable[rowIndex];
            $scope.vmTable[rowIndex] = tmp;
            $scope.postVmTable();
        }
    }

    // Move VmTable row down
    $scope.moveRowDown = function (rowIndex) {
        if (rowIndex < $scope.vmTable.length - 1) {
            tmp = $scope.vmTable[rowIndex + 1];
            $scope.vmTable[rowIndex + 1] = $scope.vmTable[rowIndex];
            $scope.vmTable[rowIndex] = tmp;
            $scope.postVmTable();
        }
    }

    // Delete row
    $scope.deleteRow = (i) => {
        $scope.vmTable.splice(i, 1);
        $scope.postVmTable();
    }

    //Stuff for editing vmTable info
    $scope.editingData = {};

    for (var i = 0, length = $scope.vmTable.length; i < length; i++) {
        $scope.editingData[$scope.vmTable[i]] = false;
    }

    $scope.modify = function (row) {
        $scope.editingData[row] = true;
    }

    $scope.update = function (row) {
        $scope.editingData[row] = false;
        $scope.postVmTable();
    }

    $scope.getVmInfo = (n, h, index) => {

        $http.get('http://151.155.216.36:5000/vm', {
            params: {
                host: h,
                name: n
            }
        }).then((res) => {
            // Success
            // Set options for circliful
            var options = {
                animation: 1,
                animationStep: 5,
                foregroundBorderWidth: 15,
                backgroundBorderWidth: 15,
                textSize: 28,
                textStyle: 'font-size: 12px;',
                textColor: '#666',
                percentageY: 95
            }
            var getColor = (percent) => {
                if (percent < 60) {
                    return '#37EC2D';
                }
                else if (percent >= 60 && percent < 80) {
                    return '#FFA330';
                }
                else if (percent >= 80) {
                    return '#FE3035';
                }
                return '#000';
            }

            // Remove everything from divs and fill them with circle graphs
            $("#cpu_circle_" + index).html('').circliful(Object.assign({
                text: "CPU",
                percent: res.data.cpu / 100,
                foregroundColor: getColor(res.data.cpu / 100)
            }, options));

            $("#mem_circle_" + index).html('').circliful(Object.assign({
                text: "RAM",
                percent: res.data.mem / 100,
                foregroundColor: getColor(res.data.mem / 100)
            }, options));

            $("#disk_circle_" + index).html('').circliful(Object.assign({
                text: "Disk",
                percent: res.data.disk,
                foregroundColor: getColor(res.data.disk)
            }, options));

        }, (res) => {
            // Error
            console.log(res);
            document.getElementById('collapse' + index)
                .appendChild(document.createElement("p")
                    .appendChild(document.createTextNode(res.data)))
        })
    }

    $scope.getConfigData();
    $scope.plotData();
    $scope.getVmTable();
    $scope.getNotesTable();
}

    //Calculates failure percentage
    function calcFailPercentage() {
        var failures = parseInt((document.getElementById("failed").innerHTML).replace(/,/g, ""));
        var passes = parseInt(document.getElementById("passed").innerHTML.replace(/,/g, ""));

        if (failures === 0 || passes === 0) {
            document.getElementById("failPercent").innerText = "0%";
        } else if (failures > passes) {
            var tmp = failureRate = (failures / passes) * 100;
            tmp = (failures / passes) * 100;
            tmp = tmp.toFixed(2);
            document.getElementById("failPercent").innerText = tmp + "%";
        } else if (failures === passes) {
            document.getElementById("failPercent").innerText = ".50%";
        } else {
            var failureRate = (failures / passes);
            var numZeros = -Math.floor(Math.log(failureRate) / Math.log(10) + 1);
            failureRate = failureRate.toFixed(numZeros + 2);

            document.getElementById("failPercent").innerText = failureRate + "%";
        }
    }

    //Calculates average login attempts per second
    function calcAvgPerSec() {
        var hms = document.getElementById("hms").innerHTML;
        var days = document.getElementById("days").innerHTML;
        var a = days * 24 * 60 * 60;
        var b = hms.split(":");
        var seconds = (+a + (+b[0]) * 60 * 60 + (+b[1]) * 60 + (+b[2]));

        var attempts = parseInt(document.getElementById("passed").innerHTML.replace(/,/g, "")) + parseInt(document.getElementById("failed").innerHTML.replace(/,/g, ""));

        var avgPerSec = (attempts / seconds).toFixed(2);

        document.getElementById("average").innerHTML = avgPerSec;
    }