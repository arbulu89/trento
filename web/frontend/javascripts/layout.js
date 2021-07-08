// ajax operation to reload the table on pagination operations
// query can be `page` or `per_page`
function reloadTable(path) {
  $.get(path, function(response, status, xhr){
    console.log(response)
    console.log(status)
    console.log(xhr)
    var table = $(response).find('.table-responsive');
    $('.table-responsive').replaceWith(table);
    var nav = $(response).find('.pagination-wrap');
    $('.pagination-wrap').replaceWith(nav);
  });
}

function blockingQuery(path, last_index) {
  $.get({url: path, headers: {'waitIndex': last_index}}, function(response, status, xhr){
    last_index = xhr.getResponseHeader("lastIndex");
    var table = $(response).find('.table-responsive');
    $('.table-responsive').replaceWith(table);
    var nav = $(response).find('.pagination-wrap');
    $('.pagination-wrap').replaceWith(nav);
  }).done(function() {
    blockingQuery(path, last_index);
  });
}

$(document).ready(function() {
  // enable bootstrap tooltips
  $('[data-toggle="tooltip"]').tooltip();

  // pagination events
  $('body').on('click', '.page-item', function() {
    var href = new URL(window.location.href);
    href.searchParams.set('page', this.value);
    path = href.pathname + href.search;
    reloadTable(path)
    history.pushState(undefined, '', href);
  });

  $('body').on('click', '.pagination-wrap .dropdown-item', function() {
    var href = new URL(window.location.href);
    href.searchParams.set('per_page', this.textContent);
    path = href.pathname + href.search;
    reloadTable(path)
    history.pushState(undefined, '', href);
  });

  $('body').on('change', '.selectpicker', function() {
    var href = new URL(window.location.href);
    href.searchParams.delete(this.name)
    values = $(this).val();
    for (let i in values) {
      if (values[i] != "") {
        href.searchParams.append(this.name, values[i]);
      }
    }

    path = href.pathname + href.search;
    reloadTable(path)
    history.pushState(undefined, '', href);
  });

  var href = new URL(window.location.href);
  var path = href.pathname + href.search;
  blockingQuery(path, 1);

});
