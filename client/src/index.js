import Chart from 'chart.js';

window.onload = () => {
    var ctx = document.getElementById('myChart');
    var myChart = new Chart(ctx, {
        type: 'bubble',
        data: {
            datasets: [{
                label: 'Pixel 250',
                data: [{
                    x: 400,
                    y: 1200,
                    r: 20
                }, {
                    x: 250,
                    y: 800,
                    r: 40
                }, {
                    x: 600,
                    y: 890,
                    r: 8
                }]
            }]
        },
        options: {
            responsive: true,
            aspectRatio: 0.5,
            scales: {
                xAxes: [{
                    type: 'linear',
                    position: 'bottom',
                    ticks: {
                        min: 0,
                        max: 1000
                    }
                }],
                yAxes: [{
                    type: 'linear',
                    position: 'left',
                    ticks: {
                        min: 0,
                        max: 2000
                    }
                }]
            }
        }
    });
}
