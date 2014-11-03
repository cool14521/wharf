/*
 *  Document   : auth.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description: 
 *
 */

'use strict';

//Auth Page Module
angular.module('auth', ['ngRoute', 'ngMessages'])
//
.controller('SigninCtrl', ['$scope', function ($scope) {
  $scope.submit = function () {
    if($scope.loginForm.$valid) {
      console.log($scope.user)
    }
  // body...
  }
}])
.controller('SignupCtrl', ['$scope', function ($scope) {

}])
.controller('ForgotCtrl', ['$scope', function ($scope) {

}])
//routes
.config(function($routeProvider, $locationProvider){
  $routeProvider
  .when('/', {
    templateUrl: 'static/views/auth/signin.html',
    controller: 'SigninCtrl'
  })
  .when('/signup', {
    templateUrl: 'static/views/auth/signup.html',
    controller: 'SignupCtrl'
  })
  .when('/forgot', {
    templateUrl: 'static/views/auth/forgot.html',
    controller: 'ForgotCtrl'
  });
})
//directive
.directive('emailValidator', [function () {
  var EMAIL_REGEXP = /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;

  return {
    require: 'ngModel',
    restrict: '',
    link: function($scope, element, attrs, ngModel) {
      ngModel.$validators.email = function(value) {
        return EMAIL_REGEXP.test(value);
      };
    }
  };
}]);
