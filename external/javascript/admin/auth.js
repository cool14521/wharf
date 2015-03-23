/*
 *  Document   : auth.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description:
 *
 */

'use strict';

//Auth Page Module
angular.module('auth', ['ngRoute', 'ngMessages', 'ngCookies', 'angular-growl'])
//Controllers
.controller('SigninCtrl', ['$scope', '$cookies', '$http', 'growl', '$window', '$timeout', function($scope, $cookies, $http, growl, $window, $timeout) {
  $scope.submitting = false;

  $scope.submit = function() {
    if ($scope.loginForm.$valid) {
      $scope.submitting = true;

      $http.post('/a1/signin', $scope.user)
        .success(function(data, status, headers, config) {
          $scope.submitting = true;
          growl.info(data.message);
          $timeout(function() {
            $window.location.href = '/dashboard';
          }, 3000);
        })
      .error(function(data, status, headers, config) {
        $scope.submitting = false;
        growl.error(data.message);
        $timeout(function() {
          $window.location.href = '/auth';
        }, 3000);
      });
    }
  }
}])
//routes
.config(function($routeProvider, $locationProvider) {
  $routeProvider
    .when('/', {
      templateUrl: '/static/views/admin/signin.html',
      controller: 'SigninCtrl'
    })
  .when('/auth', {
    templateUrl: '/static/views/admin/signin.html',
    controller: 'SigninCtrl'
  });
})
.directive('usernameValidator', [function() {
  var USERNAME_REGEXP = /^([a-z0-9_]{6,30})$/;

  return {
    require: 'ngModel',
    restrict: '',
    link: function(scope, element, attrs, ngModel) {
      ngModel.$validators.usernames = function(value) {
        return USERNAME_REGEXP.test(value);
      }
    }
  };

}])
.directive('confirmValidator', [function() {
  return {
    require: 'ngModel',
    restrict: '',
    scope: {
      passwd: "=confirmData"
    },
    link: function(scope, element, attrs, ngModel) {
      ngModel.$validators.repeat = function(value) {
        return value == scope.passwd;
      };

      scope.$watch('passwd', function() {
        ngModel.$validate();
      });
    }
  };
}])
.directive('emailValidator', [function() {
  var EMAIL_REGEXP = /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;

  return {
    require: 'ngModel',
    restrict: '',
    link: function(scope, element, attrs, ngModel) {
      ngModel.$validators.emails = function(value) {
        return EMAIL_REGEXP.test(value);
      }
    }
  };
}])
.directive('passwdValidator', [function() {
  var NUMBER_REGEXP = /\d/;
  var LETTER_REGEXP = /[A-z]/;

  return {
    require: 'ngModel',
    restrict: '',
    link: function(scope, element, attrs, ngModel) {
      ngModel.$validators.passwd = function(value) {
        return NUMBER_REGEXP.test(value) && LETTER_REGEXP.test(value);
      }
    }
  };
}]);
