module Types exposing (..)


-- External Imports
import Http exposing (Error)


-- Internal Imports


-- Types


type alias Model =
  { movies : List ApiResponse
  , movie : ApiResponse
  , route : Route
  }

type alias MovieResponse =
  { imdbID : String
  , poster : String
  , title : String
  , kind : String
  , year : String
  }

type alias ApiResponse =
  { title : String
  , fullPath : String
  , apiMovie : MovieResponse
  }

type Msg
  = MoviesFetchSucceed (List ApiResponse)
  | MovieFetchSucceed ApiResponse
  | FetchFail Http.Error

type Route
  = Movies
  | Movie String
  | TopLevel
  | NotFound

