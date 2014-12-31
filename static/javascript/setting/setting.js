/*
 *  Document   : setting.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description: 
 *
 */

'use strict';

//Auth Page Module
angular.module('setting', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl'])
.controller('SettingProfileCtrl', ['$scope', '$cookies', '$http', 'growl', '$location', '$timeout', function ($scope, $cookies, $http, growl, $location, $timeout) {
  
}])
//routes
.config(function($routeProvider, $locationProvider){
  $routeProvider
  .when('/', {
    templateUrl: 'static/views/setting/profile.html',
    controller: 'SettingProfileCtrl'
  })
  .when('/profile', {
    templateUrl: 'static/views/setting/profile.html',
    controller: 'SettingProfileCtrl'
  });
})
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