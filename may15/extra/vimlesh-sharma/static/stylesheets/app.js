$(function() {
  $( "#slider-range-max" ).slider({
      range: "max",
      min: 5,
      max: 30,
      value: 25,
      slide: function( event, ui ) {
        $( "#TileSize" ).val( ui.value );
      }
    });
    $( "#TileSize" ).val( $( "#slider-range-max" ).slider( "value" ) );  
});