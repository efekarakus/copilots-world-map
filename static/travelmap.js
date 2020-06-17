class SquareMapChart {
    create(classicalWorld, squaredWorld) {
        this.width = 1060;
        this.height = 800;
        this.isNew = true;
        this.margin = {top: 0, left: 0, right: 0, bottom: 0}
        this.world = classicalWorld;
        this.worldTile = squaredWorld;
        this.svg = d3.select('.visualization').append('svg')
            .attr("width", this.width - (this.margin.left + this.margin.right))
            .attr("height", this.height - (this.margin.top + this.margin.bottom));
        this.main = this.svg.append('g')
            .attr('transform', `translate(${this.margin.left},${this.margin.top})`);

        this.map = this.main.append('g')
            .attr('class', 'map');

        this.labels = this.main.append('g')
            .attr('class', 'labels');
    }
    update(state) {
        this.valueAccessor = state.valueAccessor;
        this.countryCodeType = state.countryCodeType;
        this.countryCodeAccessor = state.countryCodeAccessor;
        if(state.data && JSON.stringify(this.data) !== JSON.stringify(state.data)) {
            this._drawMap(state.data);
        }
        this.data = state.data;
        this.view = state.view;
        return this.svg.node();
    }
    _drawMap(data) {
        const dataDomain = d3.extent(data, d => d[this.valueAccessor]);
        const dataMapping = data.reduce( (agg, v) => {
            var country;
            if(this.countryCodeType === 'name'){
                country = this.worldTile.find( d => {
                    return d.name === v[this.countryCodeAccessor];
                });
            }

            if(country) agg[country.alpha2] = v[this.valueAccessor];
            return agg;
        }, {});
        var midPoint = dataDomain[0] + ((dataDomain[1] - dataDomain[0])/2);
        var colorScale = d3.scaleSequential()
            .domain([dataDomain[0], dataDomain[1]])
            .interpolator(d3.interpolateOrRd);
        // create a first guess for the projection
        var projection = d3.geoNaturalEarth1()
            .scale(0.9)
            .translate([0, 0]);
        // create the path
        var path = d3.geoPath()
            .projection(projection);
        // using the path determine the bounds of the current map and use
        // these to determine better values for the scale and translation
        var bounds = path.bounds(this.world);
        var scale = .95 / Math.max(
            (bounds[1][0] - bounds[0][0]) / this.width,
            (bounds[1][1] - bounds[0][1]) / this.height
        );
        var xGridDomain = d3.extent(this.worldTile, d => d.x);
        var yGridDomain = d3.extent(this.worldTile, d => d.y);
        var xWidth = this.width/(xGridDomain[1] - xGridDomain[0]);
        var yWidth = this.height/(yGridDomain[1] - yGridDomain[0]);
        var squareWidth = Math.min(xWidth, yWidth);
        var xOffset = (this.width - ((xGridDomain[1] - xGridDomain[0])*squareWidth))/2;
        var yOffset = (this.height - ((yGridDomain[1] - yGridDomain[0])*squareWidth))/2;
        var squarePosition = this.worldTile.reduce( (agg, v) => {
            agg[v.alpha2] = {
                x: (v.x * squareWidth) + xOffset,
                y: (v.y * squareWidth) + yOffset
            }
            return agg;
        }, {})
        var offset  = [
            (this.width - scale * (bounds[1][0] + bounds[0][0])) / 2,
            (this.height - scale * (bounds[1][1] + bounds[0][1])) / 2
        ];
        // new projection
        projection = d3.geoNaturalEarth1()
            .scale(scale)
            .translate(offset);
        path = path.projection(projection);
        this.squarePosition = squarePosition;
        this.squareWidth = squareWidth;
        this.path = path;
        this.labels
            .selectAll('text')
            .data(this.world.features)
            .enter().append('text')
            .text( d => d.properties.iso_a2 )
            .attr('x', d => this.squarePosition[d.properties.iso_a2].x + (this.squareWidth/2) )
            .attr('y', d => this.squarePosition[d.properties.iso_a2].y + (this.squareWidth/2) )
            .style('fill', d => {
                return dataMapping[d.properties.iso_a2] !== undefined ?
                    dataMapping[d.properties.iso_a2] >= midPoint ? '#fff' : '#000' : '#000';
            })
            .style('alignment-baseline', 'central')
            .style('text-anchor', 'middle')
            .style('font-size', '10px')
            .style('font-family', 'Helvetica, Arial, sans-serif')
            .style('opacity', 0);

        if (!this.isNew) {
            this.map.selectAll('.country')
                .style('fill', d => {
                    return dataMapping[d.properties.iso_a2] !== undefined ?
                        colorScale(dataMapping[d.properties.iso_a2]) : '#F0F0F0'
                });
            return
        }

        this.map
            .selectAll('path')
            .data(this.world.features)
            .enter().append('path')
            .attr('d', d => {
                if(this.squarePosition[d.properties.iso_a2]){
                    var x = this.squarePosition[d.properties.iso_a2].x;
                    var y = this.squarePosition[d.properties.iso_a2].y;
                    if(d.geometry.type === 'MultiPolygon'){
                        var square = [[x,y], [x+this.squareWidth,y], [x+this.squareWidth,y+this.squareWidth], [x,y+this.squareWidth], [x,y]];
                        var filteredPolygons = d.geometry.coordinates.map( coordinates => this.path({type: 'Polygon', coordinates: coordinates}));
                        return flubber.combine(filteredPolygons, square, { single: true });
                    } else {
                        return flubber.toRect(this.path(d), x, y, this.squareWidth, this.squareWidth, { maxSegmentLength: 10 });
                    }
                } else {
                    console.log('Unmatched country ' + d.properties.iso_a2);
                    return null;
                }
            })
            .style('fill', d => {
                return dataMapping[d.properties.iso_a2] !== undefined ?
                    colorScale(dataMapping[d.properties.iso_a2]) : '#F0F0F0'
            })
                .style('stroke', 'gray')
                .style('stroke-width', 1)
            .attr('class', 'country')
            .on('click', async d => {
                var fixedNames = {
                    'United States': 'United States of America',
                    'Bosnia and Herz.': 'Bosnia and Herzegovina',
                    'Trinidad & Tobago': 'Trinidad and Tobago'
                }
                var name = d.properties.name;
                name = name.replace('&', 'and');
                if (name in fixedNames) {
                    name = fixedNames[name];
                }
                await d3.json(`/visits/${name}`, {
                    method: 'POST',
                });
                var visits = await d3.json('/visits');
                var uniqueCount = await d3.json('/uniquevisits');
                this.update({
                    valueAccessor: 'Visit',
                    countryCodeType: 'name',
                    countryCodeAccessor: 'Country',
                    data: visits,
                    view: 'grid',
                });
                d3.select('.emph').text(() => uniqueCount.Count);
            })
            .transition()
            .duration(0)
            .attr('d', d => {
                if(this.squarePosition[d.properties.iso_a2]){
                    var x = this.squarePosition[d.properties.iso_a2].x;
                    var y = this.squarePosition[d.properties.iso_a2].y;
                    var square = [[x,y], [x+this.squareWidth,y], [x+this.squareWidth,y+this.squareWidth], [x,y+this.squareWidth], [x,y]];
                    return flubber.toPathString(square);
                }
            });

        this.labels
            .selectAll('text')
            .transition()
            .duration(0)
            .style('opacity', 1);

        this.isNew = false;
    }
};

async function loadData() {
    var classicalWorld = await d3.json("https://gist.githubusercontent.com/KarimDouieb/fbd29d80918c0b16aef837680eddb865/raw/0a2243629bf4c95f1853aac2aa6e2ebdda61cd38/world.geo.json");
    var countryData = await d3.json('/visits');
    var gridWorld = await d3.json("https://gist.githubusercontent.com/KarimDouieb/fbd29d80918c0b16aef837680eddb865/raw/0a2243629bf4c95f1853aac2aa6e2ebdda61cd38/worldTileGrid.json");
    var chart = new SquareMapChart();
    chart.create(classicalWorld, gridWorld);

    var state = {
        valueAccessor: 'Visit',
        countryCodeType: 'name',
        countryCodeAccessor: 'Country',
        data: countryData,
        view: 'grid',
    };
    chart.update(state);
}

loadData()