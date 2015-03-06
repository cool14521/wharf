/*
 *  Document   : admin.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description: 
 *
 */

'use strict';

//Auth Page Module
angular.module('admin', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl'])
.directive('namespaceValidator', [function (){
  var NAMESPACE_REGEXP = /^([a-z0-9_]{6,30})$/;

  return {
    require: 'ngModel',
    restrict: '',
    link: function(scope, element, attrs, ngModel) {
      ngModel.$validators.usernames = function(value) {
        return USERNAME_REGEXP.test(value);
      }
    }
  };
}]);
