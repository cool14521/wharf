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
.controller('SigninCtrl', ['$scope', '$cookies', '$http', 'growl', function ($scope, $cookies, $http, growl) {
  $http.defaults.headers.post['X-XSRFToken'] = base64_decode($cookies._xsrf.split('|')[0]);

  $scope.submitting = false;

  $scope.submit = function () {
    if($scope.loginForm.$valid) {
      $scope.submitting = true;

      $http.post('/w1/signin', $scope.user)
        .success(function(data, status, headers, config) {
          $scope.submitting = false;
           growl.info(data.message);
        })
        .error(function(data, status, headers, config) {
          $scope.submitting = false;
          growl.error(data.message);
        });
    }
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
.directive('emailValidator', [function (){
  var EMAIL_REGEXP = /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;

  return {
    require: 'ngModel',
    restrict: '',
    link: function($scope, element, attrs, ngModel) {
      ngModel.$validators.emails = function(value) {
        return EMAIL_REGEXP.test(value);
      };
    }
  };
}])
.directive('passwdValidator', [function (){
  var NUMBER_REGEXP = /\d/;
  var LETTER_REGEXP = /[A-z]/;

  return {
    require: 'ngModel',
    restrict: '',
    link: function($scope, element, attrs, ngModel) {
      ngModel.$validators.passwd = function(value) {
        return NUMBER_REGEXP.test(value) && LETTER_REGEXP.test(value);
      }
    }
  };
}]);

function base64_decode(data) {
  var b64 = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=';
  var o1, o2, o3, h1, h2, h3, h4, bits, i = 0,
      ac = 0,
      dec = '',
      tmp_arr = [];

  if (!data) {
    return data;
  }

  data += '';

  do { // unpack four hexets into three octets using index points in b64
    h1 = b64.indexOf(data.charAt(i++));
    h2 = b64.indexOf(data.charAt(i++));
    h3 = b64.indexOf(data.charAt(i++));
    h4 = b64.indexOf(data.charAt(i++));

    bits = h1 << 18 | h2 << 12 | h3 << 6 | h4;

    o1 = bits >> 16 & 0xff;
    o2 = bits >> 8 & 0xff;
    o3 = bits & 0xff;

    if (h3 == 64) {
      tmp_arr[ac++] = String.fromCharCode(o1);
    } else if (h4 == 64) {
      tmp_arr[ac++] = String.fromCharCode(o1, o2);
    } else {
      tmp_arr[ac++] = String.fromCharCode(o1, o2, o3);
    }
  } while (i < data.length);

  dec = tmp_arr.join('');

  return dec.replace(/\0+$/, '');
}
