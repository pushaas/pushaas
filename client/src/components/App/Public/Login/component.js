import React, { useContext, useState } from 'react'
import Button from '@material-ui/core/Button'
import TextField from '@material-ui/core/TextField'
import Divider from '@material-ui/core/Divider'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import Container from '@material-ui/core/Container'

import authService from 'services/authService'
import credentialsService from 'services/credentialsService'

import { useStyles } from 'components/App/Public/styles'
import SetUserContext from 'components/contexts/SetUserContext'

const useFormInput = (initial) => {
  const [value, setValue] = useState(initial)
  const handleChange = (e) => setValue(e.target.value)
  return [value, handleChange]
}

const Login = () => {
  const classes = useStyles()
  const setUser = useContext(SetUserContext)

  const [username, setUsername] = useFormInput('')
  const [password, setPassword] = useFormInput('')

  const login = (e) => {
    e.preventDefault()

    const credentials = { username, password }
    credentialsService.setCredentials(credentials)

    authService.checkAuth(credentials)
      .then(() => setUser({ username }))
  }

  return (
    <Container component="main" maxWidth="xs">
      <div className={classes.paper}>
        <Typography component="h1" variant="h5">
          PushaaS Admin
        </Typography>
        <form className={classes.form} noValidate onSubmit={login}>
          <TextField
            value={username}
            onChange={setUsername}
            variant="outlined"
            margin="normal"
            required
            fullWidth
            label="Username"
            autoComplete="email"
            autoFocus
          />
          <TextField
            value={password}
            onChange={setPassword}
            variant="outlined"
            margin="normal"
            required
            fullWidth
            label="Password"
            type="password"
            autoComplete="current-password"
          />
          <Button
            type="submit"
            fullWidth
            variant="contained"
            color="primary"
            className={classes.submit}
          >
            Sign In
          </Button>
          <Grid container>
            <Grid item xs>
              <Typography variant="body2" color="textSecondary" align="center">
                The credentials are the same Tsuru uses to call the PushaaS API
              </Typography>
            </Grid>
          </Grid>
          <Divider className={classes.infoDivider} />
          <Grid container>
            <Grid item xs>
              <Typography variant="body2" color="textSecondary" align="center">
                Please note: Currently this app uses basic auth. Credentials will be transmited in plain text to the API and stored on your browser's localStorage until you logout
              </Typography>
            </Grid>
          </Grid>
        </form>
      </div>
    </Container>
  )
}

export default Login
