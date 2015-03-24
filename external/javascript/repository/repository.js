/*
 *  Document   : repository.js
 *  Author     : Meaglith Ma <genedna@gmail.com> @genedna
 *  Description: 
 *
 */

'use strict';

//Auth Page Module
angular.module('repository', [])
//App Config
.config(['growlProvider', function(growlProvider){
  growlProvider.globalTimeToLive(3000);
}]);
