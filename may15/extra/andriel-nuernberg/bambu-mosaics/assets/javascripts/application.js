window.app = window.app || {};

(function(app, $) {
  'user strict';

  var Uploader = function(form, spinner) {
    this.form = form;
    this.spinner = spinner;
    this.inputFile = form.find('input[type="file"]');

    this.bindEvends();
  };

  Uploader.prototype.bindEvends = function() {
    this.inputFile.on('change', this.submitForm.bind(this));
  };

  Uploader.prototype.submitForm = function() {
    this.form.hide();
    this.spinner.show();
    this.form.submit();
  };

  app.Uploader = Uploader;
})(app, jQuery);
