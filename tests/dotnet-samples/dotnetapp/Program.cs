using System;
using System.Net;
using System.Runtime.InteropServices;
using static System.Console;

// Ascii text: https://ascii.co.uk/text (3D Diagonol)

WriteLine("""
                         ___                              ___     
      ,---,            ,--.'|_                          ,--.'|_   
    ,---.'|   ,---.    |  | :,'       ,---,             |  | :,'  
    |   | :  '   ,'\   :  : ' :   ,-+-. /  |            :  : ' :  
    |   | | /   /   |.;__,'  /   ,--.'|'   |   ,---.  .;__,'  /   
  ,--.__| |.   ; ,. :|  |   |   |   |  ,"' |  /     \ |  |   |    
 /   ,'   |'   | |: ::__,'| :   |   | /  | | /    /  |:__,'| :    
.   '  /  |'   | .; :  '  : |__ |   | |  | |.    ' / |  '  : |__  
'   ; |:  ||   :    |  |  | '.'||   | |  |/ '   ;   /|  |  | '.'| 
|   | '/  ' \   \  /   ;  :    ;|   | |--'  '   |  / |  ;  :    ; 
|   :    :|  `----'    |  ,   / |   |/      |   :    |  |  ,   /  
 \   \  /               ---`-'  '---'        \   \  /    ---`-'   
  `----'                                      `----'              
""");

// .NET information
WriteLine(RuntimeInformation.FrameworkDescription);

