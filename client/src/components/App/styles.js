import { makeStyles } from '@material-ui/core/styles'

export const useStyles = () => {
  return makeStyles(theme => ({
    loader: {
      alignContent: 'center',
      alignItems: 'center',
      boxSizing: 'border-box',
      display: 'flex',
      flexDirection: 'row',
      flexWrap: 'nowrap',
      justifyContent: 'center',
      height: '400px',
    }
  }))()
}
